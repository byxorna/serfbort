package main

import (
	"flag"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/serf/serf"
)

var (
	c                        = serf.DefaultConfig()
	mode          daemonMode = SlaveMode
	masterAddress string
	listenAddress string
)

type daemonMode int

const (
	MasterMode daemonMode = iota
	SlaveMode
)

func init() {
	flag.StringVar(&masterAddress, "master", "mymaster.company.net:7946", "Join the cluster by coordinating with this master")
	flag.StringVar(&listenAddress, "listen", "localhost:7946", "Listen on the address for serf communication")
	flag.Parse()
}

func main() {
	c.Tags = map[string]string{
		"role": "web",
		"env":  "dev",
	}

	fields := strings.Split(listenAddress, ":")
	if len(fields) != 2 {
		log.Fatalf("-listen requires host:port! %s is not valid", listenAddress)
	}
	bindAddr := fields[0]
	bindPort, err := strconv.Atoi(fields[1])
	if err != nil {
		log.Fatalf("Unable to parse %s into port", fields[1])
	}
	c.MemberlistConfig.BindAddr = bindAddr
	c.MemberlistConfig.BindPort = bindPort
	//TODO advertiseaddr and advertiseport
	//c.MemberlistConfig.AdvertiseAddr = x
	//c.MemberlistConfig.AdvertisePort = y

	args := flag.Args()
	if len(args) > 0 && args[0] == "master" {
		mode = MasterMode
	}

	s, err := serf.Create(c)
	if err != nil {
		log.Fatalf("Error creating Serf: %s", err)
	}

	if mode == SlaveMode {
		log.Printf("Joining %s", masterAddress)
		n, err := s.Join([]string{masterAddress}, false)
		if n > 0 {
			log.Printf("joined cluster with master %s and %d nodes", masterAddress, n)
		}
		if err != nil {
			log.Fatalf("unable to join cluster with master %s: %v", masterAddress, err)
		}
	}

	members := s.Members()
	log.Printf("%d nodes currently in cluster:", len(members))
	for _, m := range members {
		/*
		   Name   string
		   Addr   net.IP
		   Port   uint16
		   Tags   map[string]string
		   Status MemberStatus
		*/
		log.Printf("  %s %s:%d %v %s", m.Name, m.Addr, m.Port, m.Tags, m.Status)
	}

	log.Print("Running...")

	//TODO this defer isnt running when ctrl-c'd
	/*
			go func() {
		    sig := <-sigChan
		        log.Printf("Cleaning up registered callback after signal %s\n", sig)
		                              os.Exit(0)
				log.Print("fuck")
				if s != nil {
					log.Print("Leaving cluster")
					err := s.Leave()
					if err != nil {
						log.Print(err)
						return
					}
					log.Print("Shutting down")
					err = s.Shutdown()
					if err != nil {
						log.Print(err)
						return
					}

				}
			}()
	*/

	select {}
}
