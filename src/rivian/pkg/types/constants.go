package types

const (
	RivianBasePath      = "https://rivian.com/api/gql"
	RivianGatewayPath  = RivianBasePath + "/gateway/graphql"
	RivianChargingPath = RivianBasePath + "/chrg/user/graphql"
	RivianOrdersPath   = RivianBasePath + "/orders/graphql"
	RivianContentPath  = RivianBasePath + "/content/graphql"
	RivianTransactionsPath = RivianBasePath + "/t2d/graphql"
	RivianAPIPath      = "https://api.rivian.com"
)

var DefaultHeaders = map[string]string{
	"User-Agent":    "RivianApp/1304 CFNetwork/1404.0.5 Darwin/22.3.0",
	"Accept":        "application/json",
	"Content-Type":  "application/json",
	"Accept-Language": "en-US",
	"Accept-Encoding": "gzip, deflate, br",
	"Apollographql-Client-Name": "com.rivian.ios.consumer-apollo-ios",
} 