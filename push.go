package main

import (
	"fmt"
	"log"

	"github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/anachronistic/apns"
)

func (c *appContext) SendNewPublicMessagePush(userIds []int64) {
	c.sendNewMessagePush(userIds, "New message!")
}

func (c *appContext) SendNewDirectMessagePush(userIds []int64) {
	c.sendNewMessagePush(userIds, "Direct message!")
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
	var tokenUser DeviceTokenUser
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
		payload.Sound = "default"

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
    friendlyName: Apple Push Services: co.getpopulr.populr
    localKeyID: B8 27 30 CE B3 1E DC FF 8C 78 17 43 D1 9D 23 A6 A8 82 94 41 
subject=/UID=co.getpopulr.populr/CN=Apple Push Services: co.getpopulr.populr/OU=8Y2GFS326V/O=Desmond McNamee/C=US
issuer=/C=US/O=Apple Inc./OU=Apple Worldwide Developer Relations/CN=Apple Worldwide Developer Relations Certification Authority
-----BEGIN CERTIFICATE-----
MIIGIDCCBQigAwIBAgIIPIakS1ueZUkwDQYJKoZIhvcNAQELBQAwgZYxCzAJBgNV
BAYTAlVTMRMwEQYDVQQKDApBcHBsZSBJbmMuMSwwKgYDVQQLDCNBcHBsZSBXb3Js
ZHdpZGUgRGV2ZWxvcGVyIFJlbGF0aW9uczFEMEIGA1UEAww7QXBwbGUgV29ybGR3
aWRlIERldmVsb3BlciBSZWxhdGlvbnMgQ2VydGlmaWNhdGlvbiBBdXRob3JpdHkw
HhcNMTcwNDIzMDkzNzM5WhcNMTgwNTIzMDkzNzM5WjCBlDEjMCEGCgmSJomT8ixk
AQEME2NvLmdldHBvcHVsci5wb3B1bHIxMTAvBgNVBAMMKEFwcGxlIFB1c2ggU2Vy
dmljZXM6IGNvLmdldHBvcHVsci5wb3B1bHIxEzARBgNVBAsMCjhZMkdGUzMyNlYx
GDAWBgNVBAoMD0Rlc21vbmQgTWNOYW1lZTELMAkGA1UEBhMCVVMwggEiMA0GCSqG
SIb3DQEBAQUAA4IBDwAwggEKAoIBAQC6yEhETyoeNCIHBQHTlHK9hVSeI7KG4OmZ
2vPjXuiVMdAguBVrWwSDB48Bo4eTbDTh24514Kee2De2ZJ4++k1H1roq1gzj7imR
SjdEstlJa62v2h34SF5SydKj5WYb90PYB1PvKQMuqh+rOpBRedPpdARhUPvmHb+1
TNsPY9CekykVQ9SLPYqswl9Yu+CmxQshO7gklNSQwWwLUDCaLw2vvKZZcgSafONY
N8xfzyWggi/lYnM51ItTWPUp3NQwUmJw2uWzXeIjHL9R90qJw3h1oY1FdcchaiP5
TaBmtRyMsyWfyGSEwXB3b8BLOowlIJWnViKYfhH0ny3yJpEO9E4VAgMBAAGjggJw
MIICbDAdBgNVHQ4EFgQUuCcwzrMe3P+MeBdD0Z0jpqiClEEwDAYDVR0TAQH/BAIw
ADAfBgNVHSMEGDAWgBSIJxcJqbYYYIvs67r2R1nFUlSjtzCCARwGA1UdIASCARMw
ggEPMIIBCwYJKoZIhvdjZAUBMIH9MIHDBggrBgEFBQcCAjCBtgyBs1JlbGlhbmNl
IG9uIHRoaXMgY2VydGlmaWNhdGUgYnkgYW55IHBhcnR5IGFzc3VtZXMgYWNjZXB0
YW5jZSBvZiB0aGUgdGhlbiBhcHBsaWNhYmxlIHN0YW5kYXJkIHRlcm1zIGFuZCBj
b25kaXRpb25zIG9mIHVzZSwgY2VydGlmaWNhdGUgcG9saWN5IGFuZCBjZXJ0aWZp
Y2F0aW9uIHByYWN0aWNlIHN0YXRlbWVudHMuMDUGCCsGAQUFBwIBFilodHRwOi8v
d3d3LmFwcGxlLmNvbS9jZXJ0aWZpY2F0ZWF1dGhvcml0eTAwBgNVHR8EKTAnMCWg
I6Ahhh9odHRwOi8vY3JsLmFwcGxlLmNvbS93d2RyY2EuY3JsMA4GA1UdDwEB/wQE
AwIHgDATBgNVHSUEDDAKBggrBgEFBQcDAjAQBgoqhkiG92NkBgMBBAIFADAQBgoq
hkiG92NkBgMCBAIFADCBgAYKKoZIhvdjZAYDBgRyMHAME2NvLmdldHBvcHVsci5w
b3B1bHIwBQwDYXBwDBhjby5nZXRwb3B1bHIucG9wdWxyLnZvaXAwBgwEdm9pcAwg
Y28uZ2V0cG9wdWxyLnBvcHVsci5jb21wbGljYXRpb24wDgwMY29tcGxpY2F0aW9u
MA0GCSqGSIb3DQEBCwUAA4IBAQDJGNM3Ep/F+TUEWSTW3MTrprrtf0+++mSfMrqO
jcGwyxiiqk0zcInhun2b2898Oxp8KWrOZPmat/Qhl1y4Dq3SJk66SRrB+mCq1G6I
5bbXvdflA43H/BWs1OGJ23Se4kWN/xNz+nmZu600teIgpnv8v9XdiuGXo1iOM0dt
6BDQWxgLmaOxWhB93tQShuO3prwNwxzWqxxo28yoo9GtfR7p2p7gXGyWZux0JQjA
JqHb4XIbmIHJAMULESiNJGlIRQRELnvAl0bkMvhMtv1GY+FF+XlEEtmMdrhV2f2w
FqCFBN7SZlAzGGE3sbuu7r8B/jUOH4DJ6T9+rhveWBcxYntA
-----END CERTIFICATE-----
Bag Attributes
    friendlyName: Populr Push
    localKeyID: B8 27 30 CE B3 1E DC FF 8C 78 17 43 D1 9D 23 A6 A8 82 94 41 
Key Attributes: <No Attributes>
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAushIRE8qHjQiBwUB05RyvYVUniOyhuDpmdrz417olTHQILgV
a1sEgwePAaOHk2w04duOdeCnntg3tmSePvpNR9a6KtYM4+4pkUo3RLLZSWutr9od
+EheUsnSo+VmG/dD2AdT7ykDLqofqzqQUXnT6XQEYVD75h2/tUzbD2PQnpMpFUPU
iz2KrMJfWLvgpsULITu4JJTUkMFsC1Awmi8Nr7ymWXIEmnzjWDfMX88loIIv5WJz
OdSLU1j1KdzUMFJicNrls13iIxy/UfdKicN4daGNRXXHIWoj+U2gZrUcjLMln8hk
hMFwd2/ASzqMJSCVp1YimH4R9J8t8iaRDvROFQIDAQABAoIBAArwO7s0X6URLIT4
uBip8uZCbqgsMwJPHZ8TAYFpR3mlRykDXs3AsMzznl5cM05g4d/wObGRjH9y7iBS
WCocAnaYjqJ/kpWuluSZUg9F4g+4rJ1FyseFhXCXvSw3/PjaTDDUjQfOgQ80i1I9
xcHfvpdHYhJI6deaGmYFsDaAg/ElV380xKCgUU6z7GZtBwGR5u5S1a6acJeqnbYO
+C/JNLb3n960ydQ6tXVQZ/7U2f6sl89ree3ZVxqNR0eyNcf0i4C0GgJrcn79fWat
EyuLh5lk/CMDBzbJHmVeaEF2XYdwoQ7zPkCDWvBr5QGN57iyouTESqyO4TdZgMp9
XTGdoUECgYEA3etOUx2UOKZVsfbgAJib7Z+Ic0CauM1KjEpMhgxChTpaHSquECbu
fzcukM8sqsajIOct5h1rVbSbOVPGQpc2HkrjGV+gpGmdZoirT8TO1qKCZ0CrZAuz
YMOXNRioLbB2gcKIkdL6cqp3Vv7QfZHW6Ahpmu23XJxAO6as9KVe7okCgYEA13eU
7tvsKN50eXcbrf1KkOKQ+/TpOE39QI81rQnlS7+q6vHjzW4NAuB6Z4Feyh+8NUdb
jFsGj6yW7LrocjzGH8eLj5gT5FDG9UG6UF9jritlN1x97qL3HgZNDlTyL7UBgmiA
cXXFsaR+YsxrkpqZ4WvAaTHPQt7K/k0y7bkJYC0CgYBCnNct67sKJiOi/8/NXgGw
GisDLWlD/5tY8RR3SEbPZuyVxLHq90SvuwinPwjRWj6tKbeFU19copHVa2gfpPQB
s8jnXOUDdRBiLrP9hb3wf2dVRvwrU7fMW+mPbo9M56Mq4BHOc93pfXHFE0fR6Wzw
yVWpw6E+k0hUn3tbFCiiwQKBgB54dsNgroEJFIeo5G0yiLz8jWxUMjcYMFxU5E5Y
O+j+bflTw9dlXMmvXSAOF42V91PBh5zNspvW2HEZ7Y1aMtqDqaTg6M887SX56ZM8
KiiTUnkFx3lb6n6AfZ0tPiKpAlGi3act9IsurADkz8Gnw2MxjcBSnvDh+OsFl9Iu
fLjxAoGBAINOO83zdPnlDn7kEgnDuvr0EGPEw0bfd+cMm2ShkMhJhhhLdcIramdg
kk/GaG78eeFc+Bs6oUGZeM1UWJKRyAajjlK+QK/j4Vw7pKEf7w1112jeTRKy44mW
LvCyvFlF1Q875iwlcC6h8Jk32rFshewBGNTisVg1w5GRox+D8EfD
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
