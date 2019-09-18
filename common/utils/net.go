/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"net"
	"strings"
)

import (
	perrors "github.com/pkg/errors"
)

var (
	privateBlocks []*net.IPNet
)

func init() {
	for _, b := range []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"} {
		if _, block, err := net.ParseCIDR(b); err == nil {
			privateBlocks = append(privateBlocks, block)
		}
	}
}

// ref: https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
func GetLocalIP() (string, error) {
	faces, err := net.Interfaces()
	if err != nil {
		return "", perrors.WithStack(err)
	}

	var privateIpv4Addr, ipv4Addr net.IP
	for _, face := range faces {
		if face.Flags&net.FlagUp == 0 {
			// interface down
			continue
		}

		if face.Flags&net.FlagLoopback != 0 {
			// loopback interface
			continue
		}

		if strings.Contains(strings.ToLower(face.Name), "docker") {
			continue
		}

		addrs, err := face.Addrs()
		if err != nil {
			return "", perrors.WithStack(err)
		}

		if ipv4, ok := getValidIPv4(addrs); ok {
			ipv4Addr = ipv4
			if isPrivateIP(ipv4) {
				privateIpv4Addr = ipv4
			}
		}
	}

	if ipv4Addr == nil {
		return "", perrors.Errorf("can not get local IP")
	}

	if privateIpv4Addr == nil {
		return ipv4Addr.String(), nil
	}

	return privateIpv4Addr.String(), nil
}

func isPrivateIP(ip net.IP) bool {
	for _, priv := range privateBlocks {
		if priv.Contains(ip) {
			return true
		}
	}
	return false
}

func getValidIPv4(addrs []net.Addr) (net.IP, bool) {
	for _, addr := range addrs {
		var ip net.IP

		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}

		if ip == nil || ip.IsLoopback() {
			continue
		}

		ip = ip.To4()
		if ip == nil {
			// not an valid ipv4 address
			continue
		}

		return ip, true
	}
	return nil, false
}
