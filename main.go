package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type node struct {
	Ipaddr  string `json:"ip"`
	Avgping string `json:"avgping"`
}

func main() {
	ips := strings.Split(readFromIguana(), "\n")

	var nodes []*node

	for _, ip := range ips {
		if ip != "" {
			nodes = append(nodes, &node{Ipaddr: ip, Avgping: strings.Split(avgPing(ip), "\n")[0]})
		}
	}

	for _, n := range nodes {
		fmt.Printf("IP address: %s\tAverage ping: %v\n", n.Ipaddr, n.Avgping)
	}

	b, err := json.Marshal(nodes)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(b))
}

func readFromIguana() string {
	curl := exec.Command("curl", "-s", "--url", "http://127.0.0.1:7776/", "--data", "{\"agent\":\"dpow\",\"method\":\"ipaddrs\"}")
	jq := exec.Command("jq", "-r", `.[]`)
	jq.Stdin, _ = curl.StdoutPipe()
	//jq.Stdout = os.Stdout
	var outb, errb bytes.Buffer
	jq.Stdout = &outb
	jq.Stderr = &errb

	// add error checking and/or if iguana conns == 0
	_ = jq.Start()
	_ = curl.Run()
	_ = jq.Wait()

	stdout := outb.String()
	//stderr := errb.String()

	return stdout
}

func avgPing(ip string) string {
	ping := exec.Command("ping", "-c", "4", ip)
	tail := exec.Command("tail", "-1")
	awk := exec.Command("awk", "-F", "/", "{print $5}")
	tail.Stdin, _ = ping.StdoutPipe()
	awk.Stdin, _ = tail.StdoutPipe()

	var outb, errb bytes.Buffer
	awk.Stdout = &outb
	awk.Stderr = &errb

	_ = tail.Start()
	_ = awk.Start()
	_ = ping.Run()
	_ = tail.Wait()
	_ = awk.Wait()

	stdout := outb.String()

	return stdout
}

/*
func runCmd(cmdName string, arg ...string) {
	cmd := exec.Command(cmdName, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to run command. %s\n", err)
		os.Exit(1)
	}
}*/
