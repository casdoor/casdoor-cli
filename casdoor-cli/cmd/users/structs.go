/*
Copyright © 2024 Fabien
*/

package users

import "time"

// Account auth variables
type Account struct {
	AccessToken     string
	CasdoorEndpoint string
}

// GlobalUsersResponse
// GET /api/get-global-users
type GlobalUsersResponse struct {
	Status string            `json:"status"`
	Msg    string            `json:"msg"`
	Sub    string            `json:"sub"`
	Name   string            `json:"name"`
	Data   []GlobalUsersData `json:"data"`
	Data2  any               `json:"data2"`
}

type GlobalUsersData struct {
	Owner               string    `json:"owner"`
	Name                string    `json:"name"`
	CreatedTime         time.Time `json:"createdTime"`
	UpdatedTime         time.Time `json:"updatedTime"`
	DeletedTime         string    `json:"deletedTime"`
	ID                  string    `json:"id"`
	ExternalID          string    `json:"externalId"`
	Type                string    `json:"type"`
	Password            string    `json:"password"`
	PasswordSalt        string    `json:"passwordSalt"`
	PasswordType        string    `json:"passwordType"`
	DisplayName         string    `json:"displayName"`
	FirstName           string    `json:"firstName"`
	LastName            string    `json:"lastName"`
	Avatar              string    `json:"avatar"`
	AvatarType          string    `json:"avatarType"`
	PermanentAvatar     string    `json:"permanentAvatar"`
	Email               string    `json:"email"`
	EmailVerified       bool      `json:"emailVerified"`
	Phone               string    `json:"phone"`
	CountryCode         string    `json:"countryCode"`
	Region              string    `json:"region"`
	Location            string    `json:"location"`
	Address             []any     `json:"address"`
	Affiliation         string    `json:"affiliation"`
	Title               string    `json:"title"`
	IDCardType          string    `json:"idCardType"`
	IDCard              string    `json:"idCard"`
	Homepage            string    `json:"homepage"`
	Bio                 string    `json:"bio"`
	Tag                 string    `json:"tag"`
	Language            string    `json:"language"`
	Gender              string    `json:"gender"`
	Birthday            string    `json:"birthday"`
	Education           string    `json:"education"`
	Score               int       `json:"score"`
	Karma               int       `json:"karma"`
	Ranking             int       `json:"ranking"`
	Balance             int       `json:"balance"`
	Currency            string    `json:"currency"`
	IsDefaultAvatar     bool      `json:"isDefaultAvatar"`
	IsOnline            bool      `json:"isOnline"`
	IsAdmin             bool      `json:"isAdmin"`
	IsForbidden         bool      `json:"isForbidden"`
	IsDeleted           bool      `json:"isDeleted"`
	SignupApplication   string    `json:"signupApplication"`
	Hash                string    `json:"hash"`
	PreHash             string    `json:"preHash"`
	AccessKey           string    `json:"accessKey"`
	AccessSecret        string    `json:"accessSecret"`
	CreatedIP           string    `json:"createdIp"`
	LastSigninTime      string    `json:"lastSigninTime"`
	LastSigninIP        string    `json:"lastSigninIp"`
	Github              string    `json:"github"`
	Google              string    `json:"google"`
	Qq                  string    `json:"qq"`
	Wechat              string    `json:"wechat"`
	Facebook            string    `json:"facebook"`
	Dingtalk            string    `json:"dingtalk"`
	Weibo               string    `json:"weibo"`
	Gitee               string    `json:"gitee"`
	Linkedin            string    `json:"linkedin"`
	Wecom               string    `json:"wecom"`
	Lark                string    `json:"lark"`
	Gitlab              string    `json:"gitlab"`
	Adfs                string    `json:"adfs"`
	Baidu               string    `json:"baidu"`
	Alipay              string    `json:"alipay"`
	Casdoor             string    `json:"casdoor"`
	Infoflow            string    `json:"infoflow"`
	Apple               string    `json:"apple"`
	Azuread             string    `json:"azuread"`
	Azureadb2C          string    `json:"azureadb2c"`
	Slack               string    `json:"slack"`
	Steam               string    `json:"steam"`
	Bilibili            string    `json:"bilibili"`
	Okta                string    `json:"okta"`
	Douyin              string    `json:"douyin"`
	Line                string    `json:"line"`
	Amazon              string    `json:"amazon"`
	Auth0               string    `json:"auth0"`
	Battlenet           string    `json:"battlenet"`
	Bitbucket           string    `json:"bitbucket"`
	Box                 string    `json:"box"`
	Cloudfoundry        string    `json:"cloudfoundry"`
	Dailymotion         string    `json:"dailymotion"`
	Deezer              string    `json:"deezer"`
	Digitalocean        string    `json:"digitalocean"`
	Discord             string    `json:"discord"`
	Dropbox             string    `json:"dropbox"`
	Eveonline           string    `json:"eveonline"`
	Fitbit              string    `json:"fitbit"`
	Gitea               string    `json:"gitea"`
	Heroku              string    `json:"heroku"`
	Influxcloud         string    `json:"influxcloud"`
	Instagram           string    `json:"instagram"`
	Intercom            string    `json:"intercom"`
	Kakao               string    `json:"kakao"`
	Lastfm              string    `json:"lastfm"`
	Mailru              string    `json:"mailru"`
	Meetup              string    `json:"meetup"`
	Microsoftonline     string    `json:"microsoftonline"`
	Naver               string    `json:"naver"`
	Nextcloud           string    `json:"nextcloud"`
	Onedrive            string    `json:"onedrive"`
	Oura                string    `json:"oura"`
	Patreon             string    `json:"patreon"`
	Paypal              string    `json:"paypal"`
	Salesforce          string    `json:"salesforce"`
	Shopify             string    `json:"shopify"`
	Soundcloud          string    `json:"soundcloud"`
	Spotify             string    `json:"spotify"`
	Strava              string    `json:"strava"`
	Stripe              string    `json:"stripe"`
	Tiktok              string    `json:"tiktok"`
	Tumblr              string    `json:"tumblr"`
	Twitch              string    `json:"twitch"`
	Twitter             string    `json:"twitter"`
	Typetalk            string    `json:"typetalk"`
	Uber                string    `json:"uber"`
	Vk                  string    `json:"vk"`
	Wepay               string    `json:"wepay"`
	Xero                string    `json:"xero"`
	Yahoo               string    `json:"yahoo"`
	Yammer              string    `json:"yammer"`
	Yandex              string    `json:"yandex"`
	Zoom                string    `json:"zoom"`
	Metamask            string    `json:"metamask"`
	Web3Onboard         string    `json:"web3onboard"`
	Custom              string    `json:"custom"`
	WebauthnCredentials any       `json:"webauthnCredentials"`
	PreferredMfaType    string    `json:"preferredMfaType"`
	RecoveryCodes       any       `json:"recoveryCodes"`
	TotpSecret          string    `json:"totpSecret"`
	MfaPhoneEnabled     bool      `json:"mfaPhoneEnabled"`
	MfaEmailEnabled     bool      `json:"mfaEmailEnabled"`
	Invitation          string    `json:"invitation"`
	InvitationCode      string    `json:"invitationCode"`
	Ldap                string    `json:"ldap"`
	Properties          struct {
	} `json:"properties"`
	Roles               any    `json:"roles"`
	Permissions         any    `json:"permissions"`
	Groups              []any  `json:"groups"`
	LastSigninWrongTime string `json:"lastSigninWrongTime"`
	SigninWrongTimes    int    `json:"signinWrongTimes"`
	ManagedAccounts     any    `json:"managedAccounts"`
}