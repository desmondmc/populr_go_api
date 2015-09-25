package main

import (
	"fmt"
	"log"

	"github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/anachronistic/apns"
)

func (c *appContext) SendNewPublicMessagePush(userIds []int64) {
	c.sendNewMessagePush(userIds, "New message")
}

func (c *appContext) SendNewDirectMessagePush(userIds []int64) {
	c.sendNewMessagePush(userIds, "New direct message")
}

func (c *appContext) SendNewFriendPush(toUserId, fromUserId string) {
	var fromUser User
	err := c.db.Get(&fromUser, "SELECT id, username FROM users WHERE id=$1", fromUserId)
	var fromUsername string
	if err != nil {
		fromUsername = "Someone"
	} else {
		fromUsername = fromUser.Username
	}
	c.sendPushWithIdAndMessage(toUserId, "", "new_friend", fromUsername)
}

func (c *appContext) sendNewMessagePush(userIds []int64, message string) {
	for _, userId := range userIds {
		idString := fmt.Sprintf("%d", userId)
		c.sendPushWithIdAndMessage(idString, message, "new_message", "")
	}
}

func (c *appContext) sendPushWithIdAndMessage(id, message, mtype, context string) {
	var tokenUser TokenUser
	err := c.db.Get(&tokenUser, "SELECT id, username, device_token FROM users WHERE id=$1", id)
	if err != nil {
		log.Println("Error finding user for push: ", err)
	}

	log.Println("Sending push to user: ", tokenUser)

	if mtype == "new_friend" {
		message = fmt.Sprintf("%s is your friend!", context)
	}

	if tokenUser.Token != "" {
		sendPush(
			tokenUser.Token,
			message,
			mtype,
		)
	}
}

func sendPush(token, message, mtype string) {
	go func() {
		payload := apns.NewPayload()
		payload.Alert = message
		payload.Badge = 1

		pn := apns.NewPushNotification()
		pn.DeviceToken = token
		pn.AddPayload(payload)

		pn.Set("type", mtype)

		client := apns.BareClient(PRO_SERVER, CERT_PEM_RAW, KEY_PEM_RAW)
		resp := client.Send(pn)

		alert, _ := pn.PayloadString()
		log.Println("Sending Push: ")
		log.Println("  Alert:", alert)
		log.Println("Success:", resp.Success)
		log.Println("  Error:", resp.Error)
	}()
}

const CERT_PEM = "/home/dokku/populr_go_api/pop-prod-cert.pem"

const KEY_PEM = "/home/dokku/populr_go_api/pop-prod-key-noenc.pem"

const CERT_PEM_RAW = `
Bag Attributes
    friendlyName: Apple Production IOS Push Services: co.getpopulr.populr
    localKeyID: 1D 70 35 AC 0F 3C D4 9A 99 26 F6 6A 5C CC 3D B2 74 BB 6D 3C 
subject=/UID=co.getpopulr.populr/CN=Apple Production IOS Push Services: co.getpopulr.populr/OU=8Y2GFS326V/C=US
issuer=/C=US/O=Apple Inc./OU=Apple Worldwide Developer Relations/CN=Apple Worldwide Developer Relations Certification Authority
-----BEGIN CERTIFICATE-----
MIIFijCCBHKgAwIBAgIIfLN3IEnHbLwwDQYJKoZIhvcNAQEFBQAwgZYxCzAJBgNV
BAYTAlVTMRMwEQYDVQQKDApBcHBsZSBJbmMuMSwwKgYDVQQLDCNBcHBsZSBXb3Js
ZHdpZGUgRGV2ZWxvcGVyIFJlbGF0aW9uczFEMEIGA1UEAww7QXBwbGUgV29ybGR3
aWRlIERldmVsb3BlciBSZWxhdGlvbnMgQ2VydGlmaWNhdGlvbiBBdXRob3JpdHkw
HhcNMTUwNjI0MTkxMjM5WhcNMTYwNjIzMTkxMjM5WjCBiTEjMCEGCgmSJomT8ixk
AQEME2NvLmdldHBvcHVsci5wb3B1bHIxQDA+BgNVBAMMN0FwcGxlIFByb2R1Y3Rp
b24gSU9TIFB1c2ggU2VydmljZXM6IGNvLmdldHBvcHVsci5wb3B1bHIxEzARBgNV
BAsMCjhZMkdGUzMyNlYxCzAJBgNVBAYTAlVTMIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEAt50FlQ+L3iCMgJ7pO6YVnCMUlAXov8h/yZPZy4d8dPQL3xop
2aMbqM0NiaeISfPiR/lG8n4U1IP5x5CWkm2JeMip0YpQvFQSk/swYarspv5Jfn5g
U0uZAIUWw9WvB5Nxcppg74XLuycET+C7drj7AVm9yvWYTGwNf9q0svedZB+fyndE
CjtYPs8HYyb9YMeYa9wcIaaO3Xue4kgrxAY1PZJEo35PeinhLwn+bjsGbQIpkcnb
nlM5AvAIyQzp8Y8qmFt2B5fDbMETMh9aON7O7733FwcopgLR7Ur93x8odMZtbYIK
ek3CZ5zjgDX9i82hG40bzQMGDOHVRa4u6or0wQIDAQABo4IB5TCCAeEwHQYDVR0O
BBYEFB1wNawPPNSamSb2alzMPbJ0u208MAkGA1UdEwQCMAAwHwYDVR0jBBgwFoAU
iCcXCam2GGCL7Ou69kdZxVJUo7cwggEPBgNVHSAEggEGMIIBAjCB/wYJKoZIhvdj
ZAUBMIHxMIHDBggrBgEFBQcCAjCBtgyBs1JlbGlhbmNlIG9uIHRoaXMgY2VydGlm
aWNhdGUgYnkgYW55IHBhcnR5IGFzc3VtZXMgYWNjZXB0YW5jZSBvZiB0aGUgdGhl
biBhcHBsaWNhYmxlIHN0YW5kYXJkIHRlcm1zIGFuZCBjb25kaXRpb25zIG9mIHVz
ZSwgY2VydGlmaWNhdGUgcG9saWN5IGFuZCBjZXJ0aWZpY2F0aW9uIHByYWN0aWNl
IHN0YXRlbWVudHMuMCkGCCsGAQUFBwIBFh1odHRwOi8vd3d3LmFwcGxlLmNvbS9h
cHBsZWNhLzBNBgNVHR8ERjBEMEKgQKA+hjxodHRwOi8vZGV2ZWxvcGVyLmFwcGxl
LmNvbS9jZXJ0aWZpY2F0aW9uYXV0aG9yaXR5L3d3ZHJjYS5jcmwwCwYDVR0PBAQD
AgeAMBMGA1UdJQQMMAoGCCsGAQUFBwMCMBAGCiqGSIb3Y2QGAwIEAgUAMA0GCSqG
SIb3DQEBBQUAA4IBAQAEax+cWtAz64x5ABAdZCzT+H1PwYYuHPHFMiawtbvtp/bJ
R9fkAic+DUsg8zakq8sr0eO3ZzW4DXPJkvEBEzPUVanwZNQvKgwUxlKhWhXT6Wzu
kuhyExbBxfFHDcJtj0lU0D3E73w34VJbEgzNDV/5B8FR/GqK0CuGvmoii/ma7X6j
H+rbCNzs0MTJhkj3QBmePyfTP3FyYxlzZzQRA4qhinSgCjlcph60OxDs4T3K7kyr
kEJSEvZkOagk53yFOSJL8MP4Zej/O4J/UtX96mDvLaHTSk30r/1ZjI4O+iyNCahU
sPciABSCxNdClRGs1nHdCNMEsqMjvRc9Pm4RNAWm
-----END CERTIFICATE-----
Bag Attributes
    friendlyName: Desmond McNamee
    localKeyID: 1D 70 35 AC 0F 3C D4 9A 99 26 F6 6A 5C CC 3D B2 74 BB 6D 3C 
Key Attributes: <No Attributes>
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAt50FlQ+L3iCMgJ7pO6YVnCMUlAXov8h/yZPZy4d8dPQL3xop
2aMbqM0NiaeISfPiR/lG8n4U1IP5x5CWkm2JeMip0YpQvFQSk/swYarspv5Jfn5g
U0uZAIUWw9WvB5Nxcppg74XLuycET+C7drj7AVm9yvWYTGwNf9q0svedZB+fyndE
CjtYPs8HYyb9YMeYa9wcIaaO3Xue4kgrxAY1PZJEo35PeinhLwn+bjsGbQIpkcnb
nlM5AvAIyQzp8Y8qmFt2B5fDbMETMh9aON7O7733FwcopgLR7Ur93x8odMZtbYIK
ek3CZ5zjgDX9i82hG40bzQMGDOHVRa4u6or0wQIDAQABAoIBAFKu2vys64czkRG6
fbzmpYSSJ4ogvxbd6u2TLtiAQoNIArCMz5u535K8BASg9LaYpKVUk6ZPMIIijDBZ
4/Q8b1N8yTwa1RB5q4QH8VmJ5tesWtwjfK0FtfiN9hpp90+qDcRV0KEL1xqID2EJ
CmIEGsQY7WagAd0oK0wP6J9O8glLj8CGCTfW/9spN0QQ1eKgk+n3VZY6wtnuUddN
eBYfH3fRo2bEUUFWkRuZGwCrJXoFHxyaej/7bwP9JUpgxzon+9XnFV5sbz+6F0Sn
MNJg2r882750zTbCOr7dWvafFUNcGUIO+bJhEB+46RAXfM5+uipn4qsl+iS8+Eh1
mYG6HYECgYEA8//zgYFJDeBaLOz/zSo9v3m+nEnzpEc7uW8uZh9+wFIpOOPwCWxB
8f57pPGhIzKJ4ubfDXxqiRrLslGOh06OQSvvqoAVte81zKxieNG//bPUUqhne4Cj
T/MuXRms7XsZpMpel65NRC0hGP1NfppLpqHuO18zhM4XHbrZ822AqcUCgYEAwKTI
YH9eNAE/vqhTPNAfwR78Zs/aq6yD8Mgdu55cGv3/rPNTxeay9t+w67AKC/z2fyP/
okDPXdEizMpMnWOm/h8ol4EbpScikP6jXWYiYaUwZl0hn04utarKubgawiL0FQ7E
1GyyrdOzKW+KxPsfgB746HPXpcsMQI0jL/mBGs0CgYEAzsEKqKlRqwV5w+nLVCFA
E3jpEGwFnSWTMS1J5tU3RWYZchCLfKCUPKeERB9PiJCxkGhLh5TufWEMS/yZxoPF
POoq+pHwU6rwLo/AHVq20hbIioYSse8t83g/yDoSc2VFNIMapnLXHDcVfaRePzIl
enxqbzrbX+R2aAXF22TlinUCgYEAr7Tcal9hja9h88TVfs8ZV2yqrB+DBDgqc3Ai
5mBPA/ONwrKBQyzjKJboaF+9GigUr+wmmlfgi2JYCk+tx5P/2SKURHNmwqDKP1Nx
r5ubWlJvrb3eD84gfmQT2YbZKR72X0qELngiFLfVXsK04Gtn1NTrFCGsnDRxrWLN
qFE69zUCgYBnsTPosbZ5tZkEpOVl1eVfsbSOKq81/SiZenD2/7ZvZWmRQXTtz+n3
dfA2Ud7wNj++9ZCzo/WJjBD3R5SKHYYZhfxLjDgX4Q3/H2uZ4vOPG9uPDTg2lGlH
ws02hoB7QeKhHW1KMbuZ0mNZWoIVKZpgcW7wKq5VLfMXLXSSloNk1Q==
-----END RSA PRIVATE KEY-----
`

const KEY_PEM_RAW = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAt50FlQ+L3iCMgJ7pO6YVnCMUlAXov8h/yZPZy4d8dPQL3xop
2aMbqM0NiaeISfPiR/lG8n4U1IP5x5CWkm2JeMip0YpQvFQSk/swYarspv5Jfn5g
U0uZAIUWw9WvB5Nxcppg74XLuycET+C7drj7AVm9yvWYTGwNf9q0svedZB+fyndE
CjtYPs8HYyb9YMeYa9wcIaaO3Xue4kgrxAY1PZJEo35PeinhLwn+bjsGbQIpkcnb
nlM5AvAIyQzp8Y8qmFt2B5fDbMETMh9aON7O7733FwcopgLR7Ur93x8odMZtbYIK
ek3CZ5zjgDX9i82hG40bzQMGDOHVRa4u6or0wQIDAQABAoIBAFKu2vys64czkRG6
fbzmpYSSJ4ogvxbd6u2TLtiAQoNIArCMz5u535K8BASg9LaYpKVUk6ZPMIIijDBZ
4/Q8b1N8yTwa1RB5q4QH8VmJ5tesWtwjfK0FtfiN9hpp90+qDcRV0KEL1xqID2EJ
CmIEGsQY7WagAd0oK0wP6J9O8glLj8CGCTfW/9spN0QQ1eKgk+n3VZY6wtnuUddN
eBYfH3fRo2bEUUFWkRuZGwCrJXoFHxyaej/7bwP9JUpgxzon+9XnFV5sbz+6F0Sn
MNJg2r882750zTbCOr7dWvafFUNcGUIO+bJhEB+46RAXfM5+uipn4qsl+iS8+Eh1
mYG6HYECgYEA8//zgYFJDeBaLOz/zSo9v3m+nEnzpEc7uW8uZh9+wFIpOOPwCWxB
8f57pPGhIzKJ4ubfDXxqiRrLslGOh06OQSvvqoAVte81zKxieNG//bPUUqhne4Cj
T/MuXRms7XsZpMpel65NRC0hGP1NfppLpqHuO18zhM4XHbrZ822AqcUCgYEAwKTI
YH9eNAE/vqhTPNAfwR78Zs/aq6yD8Mgdu55cGv3/rPNTxeay9t+w67AKC/z2fyP/
okDPXdEizMpMnWOm/h8ol4EbpScikP6jXWYiYaUwZl0hn04utarKubgawiL0FQ7E
1GyyrdOzKW+KxPsfgB746HPXpcsMQI0jL/mBGs0CgYEAzsEKqKlRqwV5w+nLVCFA
E3jpEGwFnSWTMS1J5tU3RWYZchCLfKCUPKeERB9PiJCxkGhLh5TufWEMS/yZxoPF
POoq+pHwU6rwLo/AHVq20hbIioYSse8t83g/yDoSc2VFNIMapnLXHDcVfaRePzIl
enxqbzrbX+R2aAXF22TlinUCgYEAr7Tcal9hja9h88TVfs8ZV2yqrB+DBDgqc3Ai
5mBPA/ONwrKBQyzjKJboaF+9GigUr+wmmlfgi2JYCk+tx5P/2SKURHNmwqDKP1Nx
r5ubWlJvrb3eD84gfmQT2YbZKR72X0qELngiFLfVXsK04Gtn1NTrFCGsnDRxrWLN
qFE69zUCgYBnsTPosbZ5tZkEpOVl1eVfsbSOKq81/SiZenD2/7ZvZWmRQXTtz+n3
dfA2Ud7wNj++9ZCzo/WJjBD3R5SKHYYZhfxLjDgX4Q3/H2uZ4vOPG9uPDTg2lGlH
ws02hoB7QeKhHW1KMbuZ0mNZWoIVKZpgcW7wKq5VLfMXLXSSloNk1Q==
-----END RSA PRIVATE KEY-----
`

const DEV_SERVER = "gateway.sandbox.push.apple.com:2195"
const PRO_SERVER = "gateway.push.apple.com:2195"
