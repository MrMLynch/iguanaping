// Pulls IP addresses from iguanna connections and returns avg ping time
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type Node struct {
	Ipaddr  string `json:"ip"`
	Avgping string `json:"avgping"`
}

type Nodes struct {
	Origin string  `json:"origin"`
	Nodes  []*Node `json:"pingstats"`
}

func main() {
	ips := strings.Split(readFromIguana(), "\n")
	origin := originIP()

	var n Nodes

	n.Origin = origin

	for _, ip := range ips {
		if ip != "" {
			n.Nodes = append(n.Nodes, &Node{Ipaddr: ip, Avgping: strings.Split(avgPing(ip), "\n")[0]})
		}
	}

	b, err := json.Marshal(n)
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

	// TODO: add error checking and/or if iguana conns == 0
	_ = jq.Start()
	_ = curl.Run()
	_ = jq.Wait()

	stdout := outb.String()
	//stderr := errb.String()

	return stdout
}

func originIP() string {
	curl := exec.Command("curl", "icanhazip.com")
	var stdout bytes.Buffer
	curl.Stdout = &stdout
	err := curl.Run()
	if err != nil {
		fmt.Println("couldn't get your ip :(")
	}

	return strings.Split(stdout.String(), "\n")[0]
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
