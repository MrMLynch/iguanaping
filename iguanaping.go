// Pulls IP addresses from iguanna connections and returns avg ping time
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/mrmlynch/iguanaping/utils"
)

type Node struct {
	Ipaddr  string `json:"ip"`
	Avgping string `json:"avgping"`
}

type Nodes struct {
	OriginIP    string  `json:"originip"`
	NnName      string  `json:"nn_name"`
	Ntx24h      string  `json:"ntx24h"`
	AveragePing string  `json:"avgofavgping"`
	Connections string  `json:"connections"`
	Nodes       []*Node `json:"pingstats"`
}

func main() {
	var nname string
	flag.StringVar(&nname, "name", "", `
	usage: iguanaping -name=<nn_name>
	example: iguanaping -name=mrlynch`)

	flag.Parse()
	if nname == "" {
		log.Fatal(`Please provide your NN name as listed in testnet.json

		usage: iguanaping -name=<nn_name>
		example: iguanaping -name=mrlynch`)
	}

	ips := strings.Split(readFromIguana(), "\n")
	//ips := readFromIguana()
	var n Nodes

	n.NnName = nname
	n.OriginIP = originIP()

	notaCount := utils.NtxCountFunc()

	for _, op := range notaCount {
		if op.NnName == nname {
			n.Ntx24h = strconv.Itoa(op.Total)
		}
	}

	for _, ip := range ips {
		if ip != "" && ip != n.OriginIP {
			n.Nodes = append(n.Nodes, &Node{Ipaddr: ip, Avgping: strings.Split(avgPing(ip), "\n")[0]})
		}
	}

	connsping := 0.0
	sum := 0.0

	for _, avg := range n.Nodes {
		avgping, _ := strconv.ParseFloat(avg.Avgping, 64)

		if avgping != 0.0 {
			connsping++
		}
		sum += avgping
	}

	n.AveragePing = strconv.FormatFloat(sum/connsping, 'f', 4, 64)
	n.Connections = strconv.Itoa(len(n.Nodes))

	b, err := json.Marshal(n)
	if err != nil {
		fmt.Println(err)
	}

	sendJSON(string(b))
	fmt.Println(string(b))
}

func sendJSON(json string) {
	curl := exec.Command("curl", "-v", "--request", "POST", "--data", json, "http://oracle.earth/upload_json/")
	err := curl.Run()
	if err != nil {
		fmt.Println("couldn't send your json :(", err)
	}
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
	curl := exec.Command("curl", "-s4", "checkip.amazonaws.com")
	var stdout bytes.Buffer
	curl.Stdout = &stdout
	err := curl.Run()
	if err != nil {
		fmt.Println("couldn't get your ip :(", err)
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
