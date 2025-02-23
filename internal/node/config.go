package node

import "os"

func GetHttpPort() string {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	return httpPort
}

func GetMemberPort() string {
	memberPortStr := os.Getenv("MEMBERLIST_PORT")
	if memberPortStr == "" {
		memberPortStr = "7946"
	}
	return memberPortStr
}

func GetNodeName() string {
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		nodeName = "gonolith1"
	}
	return nodeName
}
