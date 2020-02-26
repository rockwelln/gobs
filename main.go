package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

func getSystemVersion(s *Session) string {
	r, err := s.SendCommand(NewSystemRequest("SystemSoftwareVersionGetRequest"))
	if err != nil {
		panic(err)
	}
	resp := <-r.Result
	v, err := resp.Get("BroadsoftDocument.command.version")
	if err != nil {
		panic(err)
	}
	return v.(string)
}

func getUsers(s *Session) []map[string]string {
	r, err := s.SendCommand(NewSystemRequest("UserGetListInSystemRequest"))
	if err != nil {
		panic(err)
	}
	resp := <-r.Result
	users, _ := resp.GetTable("BroadsoftDocument.command.userTable")
	return users
}

func getUser(s *Session, userID string) (map[string]string, error) {
	sc, err := NewSearchCriteria(SearchModeStartsWith, SearchFieldUserID, userID, true)
	if err != nil {
		return nil, err
	}
	cmd := NewUserGetListInSystemRequest()
	cmd.Criteria = append(cmd.Criteria, *sc)
	r, err := s.SendCommand(cmd)
	if err != nil {
		return nil, err
	}
	resp := <-r.Result
	users, err := resp.GetTable("BroadsoftDocument.command.userTable")
	if len(users) == 0 {
		return nil, fmt.Errorf("not user found: %s", userID)
	}
	return users[0], nil
}

func getSca(s *Session, userID string) ([]map[string]string, error) {
	r, err := s.SendCommand(NewUserGetRequest("UserSharedCallAppearanceGetRequest16sp2", userID))
	if err != nil {
		return nil, err
	}
	resp := <-r.Result
	return resp.GetTable("BroadsoftDocument.command.endpointTable")
}

func getScaEndpoints(s *Session, userID string, scaEndpointKeys [][3]string) (map[string]string, error) {
	commands := make([]BSCommand, len(scaEndpointKeys))
	for i, key := range scaEndpointKeys {
		deviceName, deviceLevel, linePort := key[0], key[1], key[2]
		commands[i] = NewUserSharedCallAppearanceGetEndpointRequest(userID, deviceName, deviceLevel, linePort)
	}
	if len(scaEndpointKeys) == 0 {
		return nil, nil
	}

	r, err := s.SendMultipleCommands(commands)
	if err != nil {
		return nil, err
	}
	resp := <-r.Result
	details, err := resp.Get("BroadsoftDocument.command")
	if err != nil {
		return nil, err
	}
	return details.(map[string]string), nil
}

type userServices struct {
	ServicePacks []map[string]string
	Services     []map[string]string
}

func getUserServices(s *Session, userID string) (*userServices, error) {
	r, err := s.SendCommand(NewUserGetRequest("UserServiceGetAssignmentListRequest", userID))
	if err != nil {
		return nil, err
	}
	resp := <-r.Result
	packs, err := resp.GetTable("BroadsoftDocument.command.servicePacksAssignmentTable")
	if err != nil {
		return nil, err
	}
	services, err := resp.GetTable("BroadsoftDocument.command.userServicesAssignmentTable")
	if err != nil {
		return nil, err
	}
	return &userServices{
		ServicePacks: packs,
		Services:     services,
	}, nil
}

func main() {
	var (
		port     int
		host     string
		username string
		password string
	)
	flag.StringVar(&host, "host", "", "BSFT host to connect to")
	flag.IntVar(&port, "port", 2208, "BSFT port for OCI-P requests")
	flag.StringVar(&username, "user", "", "username of the user session")
	flag.StringVar(&password, "pass", "", "password of the user session")
	flag.Parse()

	if host == "" || username == "" || password == "" {
		flag.Usage()
		os.Exit(1)
	}

	co := NewConnection(host, port)
	err := co.Connect()
	if err != nil {
		panic(err)
	}
	defer co.Close(true)
	fmt.Println("Connected!")
	s, err := co.StartSession(username, password)
	if err != nil {
		panic(err)
	}
	fmt.Println("logged in")
	fmt.Println("version: ", getSystemVersion(s))
	users := getUsers(s)
	fmt.Printf("%d users found\n", len(users))

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("list user [y/n]: ")
	text, _ := reader.ReadString('\n')

	if strings.ToLower(text) == "y\n" {
		for _, u := range users {
			fmt.Println(u)
		}
	}
	fmt.Print("finding details about 1 user: ")
	text, _ = reader.ReadString('\n')
	if text != "\n" {
		text = strings.Trim(text, "\n ")
		user, err := getUser(s, text)
		if err != nil {
			panic(err)
		}
		if b, err := json.MarshalIndent(user, "", "  "); err == nil {
			fmt.Println("user:", string(b))
		}

		userID := user["User Id"]
		sca, err := getSca(s, userID)
		if err != nil {
			panic(err)
		}
		if b, err := json.MarshalIndent(sca, "", "  "); err == nil {
			fmt.Println("sca:", string(b))
		}
		services, err := getUserServices(s, userID)
		if err != nil {
			panic(err)
		}
		if b, err := json.MarshalIndent(services, "", "  "); err == nil {
			fmt.Println("user services:", string(b))
		}
	} else {
		for _, u := range users {
			sca, err := getSca(s, u["User Id"])
			if err != nil {
				panic(err)
			}
			if len(sca) != 0 {
				fmt.Println("user with sca:", u["User Id"])
			}
		}
	}
	fmt.Println("closed!")
}
