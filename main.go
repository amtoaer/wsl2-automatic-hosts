package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
)

const (
	content  = "# WSL START\n%s\n# WSL END"
	hostFile = "/mnt/c/Windows/System32/drivers/etc/hosts"
)

var regex = regexp.MustCompile("# WSL START\n(.+?)\n# WSL END")

func getBindDomains() []string {
	var (
		result      []byte
		resultStr   string
		err         error
		configFiles = make([]string, 0, 2)
	)
	if home, err := homedir.Expand("~/wah/domains"); err == nil {
		configFiles = append(configFiles, home)
	}
	configFiles = append(configFiles, "/etc/wah/domains")
	for _, file := range configFiles {
		result, err = os.ReadFile(file)
		if err != nil {
			continue
		}
	}
	if err == nil {
		resultStr = string(result)
	} else {
		resultStr = "linux"
	}
	return strings.Split(resultStr, " ")
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func writeHost(domains []string, ip string) error {
	var (
		hostList []string
	)
	if ip == "" {
		return errors.New("no ip detected")
	}
	hosts, err := os.ReadFile(hostFile)
	if err != nil {
		return err
	}
	hostsStr := string(hosts)
	regexResult := regex.FindStringSubmatch(string(hosts))
	if len(regexResult) == 0 {
		// not sure if hostsStr ends with \n, add one
		hostsStr += "\n"
		hostList = make([]string, len(domains))
		for idx, domain := range domains {
			hostList[idx] = fmt.Sprintf("%s\t%s", ip, domain)
		}
	} else {
		hostsStr = strings.Replace(hostsStr, regexResult[0], "", 1)
		hostList = strings.Split(regexResult[1], "\n")
		domainMapper := make(map[string]int)
		for idx, hostItem := range hostList {
			hostItems := strings.Split(hostItem, "\t")
			if len(hostItems) != 2 {
				return errors.New("wrong hosts file format")
			}
			domainMapper[strings.TrimSpace(hostItems[1])] = idx
		}
		for _, domain := range domains {
			if idx, ok := domainMapper[domain]; ok {
				hostList[idx] = fmt.Sprintf("%s\t%s", ip, domain)
			} else {
				hostList = append(hostList, fmt.Sprintf("%s\t%s", ip, domain))
			}
		}
	}
	return os.WriteFile(
		hostFile, []byte(hostsStr+fmt.Sprintf(content, strings.Join(hostList, "\n"))), 0755,
	)
}

func main() {
	if err := writeHost(getBindDomains(), getLocalIP()); err != nil {
		fmt.Printf("exec error: %s\n", err.Error())
		return
	}
	fmt.Println("exec success")
}
