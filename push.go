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

// HELLO ME IN 2018! Follow these steps to generate the key below.
/****
- Go onto apple dev portal and create a production Populr push cert.
- Download it and add it to your keychain.
- Export it as p12
- Private Key (KEY_PEM_RAW):
	- Run 'openssl pkcs12 -in populr_push.p12 -nocerts -nodes | openssl rsa > id_rsa'
	- This might force you to set a password. If it does you'll need to recrypt it
	- Run 'openssl rsa -in privateKey.pem -out decryptedPrivateKey.pem'

- Cert (CERT_PEM_RAW):
	- Run 'openssl pkcs12 -in populr_push.p12 -out cert.pem -nodes -clcerts'
****/

const CERT_PEM_RAW = `
Bag Attributes
    friendlyName: Apple Push Services: co.getpopulr.populr
    localKeyID: 92 E0 0C DB 0C 9B A9 E7 F3 21 50 86 28 4D 67 7D 9D EB C0 60 
subject=/UID=co.getpopulr.populr/CN=Apple Push Services: co.getpopulr.populr/OU=8Y2GFS326V/O=Desmond McNamee/C=CA
issuer=/C=US/O=Apple Inc./OU=Apple Worldwide Developer Relations/CN=Apple Worldwide Developer Relations Certification Authority
-----BEGIN CERTIFICATE-----
MIIGIDCCBQigAwIBAgIIXoHUGfrunyMwDQYJKoZIhvcNAQELBQAwgZYxCzAJBgNV
BAYTAlVTMRMwEQYDVQQKDApBcHBsZSBJbmMuMSwwKgYDVQQLDCNBcHBsZSBXb3Js
ZHdpZGUgRGV2ZWxvcGVyIFJlbGF0aW9uczFEMEIGA1UEAww7QXBwbGUgV29ybGR3
aWRlIERldmVsb3BlciBSZWxhdGlvbnMgQ2VydGlmaWNhdGlvbiBBdXRob3JpdHkw
HhcNMTgwNjE1MTYxMDQ3WhcNMTkwNzE1MTYxMDQ3WjCBlDEjMCEGCgmSJomT8ixk
AQEME2NvLmdldHBvcHVsci5wb3B1bHIxMTAvBgNVBAMMKEFwcGxlIFB1c2ggU2Vy
dmljZXM6IGNvLmdldHBvcHVsci5wb3B1bHIxEzARBgNVBAsMCjhZMkdGUzMyNlYx
GDAWBgNVBAoMD0Rlc21vbmQgTWNOYW1lZTELMAkGA1UEBhMCQ0EwggEiMA0GCSqG
SIb3DQEBAQUAA4IBDwAwggEKAoIBAQDH0JvO/dATmxhX0Xbn24DL5kvzqbgOEZDe
V8cwXdurACdpGOsaOzc/UgD42Zz03fM1mUgEz6MnXC/n3qPIz2EwR+PdudON6wjy
IzH8Qv+7eSqVqNAL7uHqAfiLdugXsZ2eLl8hAArksQ4pTop9pnFB2HHZA8XjxNxT
jDIAt6I48rieAt8WwyVjkJ2dOBsUxPFZ4UntZ6k7TxeqTmg5zs4gi+sw2QWaqD0W
6ewrca6p5CwMcvYhKgGNM5dyYkdKwoLsUgBXYDKtOP+OkpoiI4uK5aNYvicqBHBP
OmGeHkOYas34t4b3nMkhrbvtA1+pLXMk6A+GRaHOfaCLcNVhRW2jAgMBAAGjggJw
MIICbDAMBgNVHRMBAf8EAjAAMB8GA1UdIwQYMBaAFIgnFwmpthhgi+zruvZHWcVS
VKO3MIIBHAYDVR0gBIIBEzCCAQ8wggELBgkqhkiG92NkBQEwgf0wgcMGCCsGAQUF
BwICMIG2DIGzUmVsaWFuY2Ugb24gdGhpcyBjZXJ0aWZpY2F0ZSBieSBhbnkgcGFy
dHkgYXNzdW1lcyBhY2NlcHRhbmNlIG9mIHRoZSB0aGVuIGFwcGxpY2FibGUgc3Rh
bmRhcmQgdGVybXMgYW5kIGNvbmRpdGlvbnMgb2YgdXNlLCBjZXJ0aWZpY2F0ZSBw
b2xpY3kgYW5kIGNlcnRpZmljYXRpb24gcHJhY3RpY2Ugc3RhdGVtZW50cy4wNQYI
KwYBBQUHAgEWKWh0dHA6Ly93d3cuYXBwbGUuY29tL2NlcnRpZmljYXRlYXV0aG9y
aXR5MBMGA1UdJQQMMAoGCCsGAQUFBwMCMDAGA1UdHwQpMCcwJaAjoCGGH2h0dHA6
Ly9jcmwuYXBwbGUuY29tL3d3ZHJjYS5jcmwwHQYDVR0OBBYEFJLgDNsMm6nn8yFQ
hihNZ32d68BgMA4GA1UdDwEB/wQEAwIHgDAQBgoqhkiG92NkBgMBBAIFADAQBgoq
hkiG92NkBgMCBAIFADCBgAYKKoZIhvdjZAYDBgRyMHAME2NvLmdldHBvcHVsci5w
b3B1bHIwBQwDYXBwDBhjby5nZXRwb3B1bHIucG9wdWxyLnZvaXAwBgwEdm9pcAwg
Y28uZ2V0cG9wdWxyLnBvcHVsci5jb21wbGljYXRpb24wDgwMY29tcGxpY2F0aW9u
MA0GCSqGSIb3DQEBCwUAA4IBAQB4yb/YOL46CpS58Xj//0gpKjzZEOlZYJFzQso5
z3z4l5R3mT7rlXrKchMSJaogKIrz86o5uDcbD3ynd8fk+6kpi7+N0buLzuEHcS0V
M/ONi70Fyhj2+JWWpEhvD5WGHi/Ebb1AKKLiOUbvbnWKDB38wgVxcBpkdKSzne2u
BgTIwrNnQ/CCSQ5cyD7WFNs2k/jbqTRCRmPsDyzcDomgy8pEFN/yoMhC1Ve2BYrS
TB+2mHVXJjmXb1L4kZjV9TGsCuaJ+tso5QPRdkkkgGuthu+8ExcYuvhXwibvFD3C
Pn41AfbSEflhQkomC5tUft1dpzHLI/34P95XnmSDIP7bj7yg
-----END CERTIFICATE-----
Bag Attributes
    friendlyName: Desmond McNamee
    localKeyID: 92 E0 0C DB 0C 9B A9 E7 F3 21 50 86 28 4D 67 7D 9D EB C0 60 
Key Attributes: <No Attributes>
-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDH0JvO/dATmxhX
0Xbn24DL5kvzqbgOEZDeV8cwXdurACdpGOsaOzc/UgD42Zz03fM1mUgEz6MnXC/n
3qPIz2EwR+PdudON6wjyIzH8Qv+7eSqVqNAL7uHqAfiLdugXsZ2eLl8hAArksQ4p
Top9pnFB2HHZA8XjxNxTjDIAt6I48rieAt8WwyVjkJ2dOBsUxPFZ4UntZ6k7Txeq
Tmg5zs4gi+sw2QWaqD0W6ewrca6p5CwMcvYhKgGNM5dyYkdKwoLsUgBXYDKtOP+O
kpoiI4uK5aNYvicqBHBPOmGeHkOYas34t4b3nMkhrbvtA1+pLXMk6A+GRaHOfaCL
cNVhRW2jAgMBAAECggEABqJmifjjb8M2i7PovsoK/Looy25XLiBnLvq7Il/7azIW
HOc25ygQLYoLfje7fmXgLWdpfL9oK2ZkMM6zoXdBQtkgN7xzycKECYYhor6WvyLG
xzFTtxG3bo2iMv8/tYmwMhBYGEBoHC/EyJpS/tkH6afGH+6AI4/lwOBmyJ9QQ5O8
h+b8nVzFWYZV42+aXNaiqyswoTMRB5re1IfsHwR+Wz3qJbBWyTohtElYK9ud6uGB
E5WAnS0cU82Ubn3g2st3TJzEhtQL99sTN1lcXt+p6IP+Bzy4otPFyiqYl1xvFErH
WRglGC+F/C4+kvxQ7srcs3WxsNH00CqBhLGLAaP8AQKBgQDmnLZhDRLLN1Un/xK1
8yRyeJQYgIhC4G+kQASG0VB4H8CMrm1OeJchfM7KiJKdOUl/Oqyby3r4nPruBCd4
n2ZEEHGEYNJyQBuTVU8WO6Bne3/vdX+caVHNjFk6OfYwByJ3/GyaEE5Woh0PvuAB
hxZUr7brDzK+u4BGXK8aLyEmtwKBgQDdz/GbpNHqX5A7/2HjquvZBVFzu38jN5cq
Eufr4sYsn1nLQxHG8uZo9ux9Qmz6FlG9KXiuubdUH/8485riWtod6JMsezEv95SX
O4x9SNx16+ErCYJSVdOBHxHFR1R+xrYGeJZuvzvWxLbvzNtgl0p5AhYu/iyTxALN
GbDTSrskdQKBgQCSinDGORWBNtcRBGAyaJ/3cbHB5CMyRAYNXHTD6sx0mNC1VL22
yKBYskOBpclsyRNwGqvGkEXDJ5W4m8EtQDUu+Tf5Q8FWwnADbolD+n3SZEMGuiZu
EOrfb9jfTCepm08G6ctlFwmAuaE3+TXFIr9I7yOQOOcpFmLL02edfudU/wKBgQCo
fEylEmRVKDPSLyG3Ity1Y4HEbDadlJthXS0Hk6E+wegeKpr1SQpVzsJCP1Ox/4Ql
MLw31F/6KbffFcOfjq1BrKkmT4lES0Z2PchwXgkAFaVa4IU6b3ESEnyYIp9/EQex
EKYMB3y3nYLr0esNir4J/tjE51MLBwetrYcQaCKRBQKBgBsGM/c9IPeG1TRHtw6O
UZKfdFFxbCX+QuKKHZudPsSSQR2OXgIKBEDEYrKciHia3vm9p8Q7s0g7R/1bhlV3
5bEfIwQ1WNBKwcQddwB9lUZwX23zk/I3ioqPEeO0RuZ+yyhwkNYhEl1tNg+sqUQ+
nR4PoTiZiK2lqv7BlqK/efb2
-----END PRIVATE KEY-----
`

const KEY_PEM_RAW = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAx9Cbzv3QE5sYV9F259uAy+ZL86m4DhGQ3lfHMF3bqwAnaRjr
Gjs3P1IA+Nmc9N3zNZlIBM+jJ1wv596jyM9hMEfj3bnTjesI8iMx/EL/u3kqlajQ
C+7h6gH4i3boF7Gdni5fIQAK5LEOKU6KfaZxQdhx2QPF48TcU4wyALeiOPK4ngLf
FsMlY5CdnTgbFMTxWeFJ7WepO08Xqk5oOc7OIIvrMNkFmqg9FunsK3GuqeQsDHL2
ISoBjTOXcmJHSsKC7FIAV2AyrTj/jpKaIiOLiuWjWL4nKgRwTzphnh5DmGrN+LeG
95zJIa277QNfqS1zJOgPhkWhzn2gi3DVYUVtowIDAQABAoIBAAaiZon442/DNouz
6L7KCvy6KMtuVy4gZy76uyJf+2syFhznNucoEC2KC343u35l4C1naXy/aCtmZDDO
s6F3QULZIDe8c8nChAmGIaK+lr8ixscxU7cRt26NojL/P7WJsDIQWBhAaBwvxMia
Uv7ZB+mnxh/ugCOP5cDgZsifUEOTvIfm/J1cxVmGVeNvmlzWoqsrMKEzEQea3tSH
7B8Efls96iWwVsk6IbRJWCvbnerhgROVgJ0tHFPNlG594NrLd0ycxIbUC/fbEzdZ
XF7fqeiD/gc8uKLTxcoqmJdcbxRKx1kYJRgvhfwuPpL8UO7K3LN1sbDR9NAqgYSx
iwGj/AECgYEA5py2YQ0SyzdVJ/8StfMkcniUGICIQuBvpEAEhtFQeB/AjK5tTniX
IXzOyoiSnTlJfzqsm8t6+Jz67gQneJ9mRBBxhGDSckAbk1VPFjugZ3t/73V/nGlR
zYxZOjn2MAcid/xsmhBOVqIdD77gAYcWVK+26w8yvruARlyvGi8hJrcCgYEA3c/x
m6TR6l+QO/9h46rr2QVRc7t/IzeXKhLn6+LGLJ9Zy0MRxvLmaPbsfUJs+hZRvSl4
rrm3VB//OPOa4lraHeiTLHsxL/eUlzuMfUjcdevhKwmCUlXTgR8RxUdUfsa2BniW
br871sS278zbYJdKeQIWLv4sk8QCzRmw00q7JHUCgYEAkopwxjkVgTbXEQRgMmif
93GxweQjMkQGDVx0w+rMdJjQtVS9tsigWLJDgaXJbMkTcBqrxpBFwyeVuJvBLUA1
Lvk3+UPBVsJwA26JQ/p90mRDBrombhDq32/Y30wnqZtPBunLZRcJgLmhN/k1xSK/
SO8jkDjnKRZiy9NnnX7nVP8CgYEAqHxMpRJkVSgz0i8htyLctWOBxGw2nZSbYV0t
B5OhPsHoHiqa9UkKVc7CQj9Tsf+EJTC8N9Rf+im33xXDn46tQaypJk+JREtGdj3I
cF4JABWlWuCFOm9xEhJ8mCKffxEHsRCmDAd8t52C69HrDYq+Cf7YxOdTCwcHra2H
EGgikQUCgYAbBjP3PSD3htU0R7cOjlGSn3RRcWwl/kLiih2bnT7EkkEdjl4CCgRA
xGKynIh4mt75vafEO7NIO0f9W4ZVd+WxHyMENVjQSsHEHXcAfZVGcF9t85PyN4qK
jxHjtEbmfssocJDWIRJdbTYPrKlEPp0eD6E4mYitpar+wZaiv3n29g==
-----END RSA PRIVATE KEY-----
`

const DEV_SERVER = "gateway.sandbox.push.apple.com:2195"
const PRO_SERVER = "gateway.push.apple.com:2195"
