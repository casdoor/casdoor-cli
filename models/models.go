package models

import "time"

type TokenData struct {
	OAuth2Token struct {
		AccessToken  string    `json:"access_token"`
		TokenType    string    `json:"token_type"`
		RefreshToken string    `json:"refresh_token"`
		Expiry       time.Time `json:"expiry"`
	} `json:"OAuth2Token"`
	IDTokenClaims struct {
		Owner             string        `json:"owner"`
		Name              string        `json:"name"`
		CreatedTime       string        `json:"createdTime"`
		UpdatedTime       string        `json:"updatedTime"`
		DeletedTime       string        `json:"deletedTime"`
		ID                string        `json:"id"`
		Type              string        `json:"type"`
		Password          string        `json:"password"`
		PasswordSalt      string        `json:"passwordSalt"`
		PasswordType      string        `json:"passwordType"`
		DisplayName       string        `json:"displayName"`
		FirstName         string        `json:"firstName"`
		LastName          string        `json:"lastName"`
		Avatar            string        `json:"avatar"`
		AvatarType        string        `json:"avatarType"`
		PermanentAvatar   string        `json:"permanentAvatar"`
		Email             string        `json:"email"`
		EmailVerified     bool          `json:"emailVerified"`
		Phone             string        `json:"phone"`
		CountryCode       string        `json:"countryCode"`
		Region            string        `json:"region"`
		Location          string        `json:"location"`
		Address           []interface{} `json:"address"`
		Affiliation       string        `json:"affiliation"`
		Title             string        `json:"title"`
		IDCardType        string        `json:"idCardType"`
		IDCard            string        `json:"idCard"`
		Homepage          string        `json:"homepage"`
		Bio               string        `json:"bio"`
		Language          string        `json:"language"`
		Gender            string        `json:"gender"`
		Birthday          string        `json:"birthday"`
		Education         string        `json:"education"`
		Score             int           `json:"score"`
		Karma             int           `json:"karma"`
		Ranking           int           `json:"ranking"`
		IsDefaultAvatar   bool          `json:"isDefaultAvatar"`
		IsOnline          bool          `json:"isOnline"`
		IsAdmin           bool          `json:"isAdmin"`
		IsForbidden       bool          `json:"isForbidden"`
		IsDeleted         bool          `json:"isDeleted"`
		SignupApplication string        `json:"signupApplication"`
		Hash              string        `json:"hash"`
		PreHash           string        `json:"preHash"`
		AccessKey         string        `json:"accessKey"`
		AccessSecret      string        `json:"accessSecret"`
		Github            string        `json:"github"`
		Google            string        `json:"google"`
		Qq                string        `json:"qq"`
		Wechat            string        `json:"wechat"`
		Facebook          string        `json:"facebook"`
		Dingtalk          string        `json:"dingtalk"`
		Weibo             string        `json:"weibo"`
		Gitee             string        `json:"gitee"`
		Linkedin          string        `json:"linkedin"`
		Wecom             string        `json:"wecom"`
		Lark              string        `json:"lark"`
		Gitlab            string        `json:"gitlab"`
		CreatedIP         string        `json:"createdIp"`
		LastSigninTime    string        `json:"lastSigninTime"`
		LastSigninIP      string        `json:"lastSigninIp"`
		PreferredMfaType  string        `json:"preferredMfaType"`
		RecoveryCodes     interface{}   `json:"recoveryCodes"`
		TotpSecret        string        `json:"totpSecret"`
		MfaPhoneEnabled   bool          `json:"mfaPhoneEnabled"`
		MfaEmailEnabled   bool          `json:"mfaEmailEnabled"`
		Ldap              string        `json:"ldap"`
		Properties        struct {
		} `json:"properties"`
		Roles []struct {
			Name string `json:"name"`
		} `json:"roles"`
		Permissions         []interface{} `json:"permissions"`
		Groups              []string      `json:"groups"`
		LastSigninWrongTime string        `json:"lastSigninWrongTime"`
		SigninWrongTimes    int           `json:"signinWrongTimes"`
		TokenType           string        `json:"tokenType"`
		Tag                 string        `json:"tag"`
		Scope               string        `json:"scope"`
		Iss                 string        `json:"iss"`
		Sub                 string        `json:"sub"`
		Aud                 []string      `json:"aud"`
		Exp                 int           `json:"exp"`
		Nbf                 int           `json:"nbf"`
		Iat                 int           `json:"iat"`
		Jti                 string        `json:"jti"`
	} `json:"IDTokenClaims"`
}

type CasdoorConfig struct {
	Endpoint         string
	ClientID         string
	ClientSecret     string
	Certificate      string
	OrganizationName string
	ApplicationName  string
	RedirectURI      string
}
