package main

import (
	"log"

	"github.com/jailtonjunior94/keycloak-sdk-go/pkg/keycloak"
)

func main() {
	keycloakSDK, err := keycloak.NewKeycloakSDK("http://localhost:8080", "admin", "admin")
	if err != nil {
		panic(err)
	}

	realm, err := keycloakSDK.CreateRealm("Realm_SDK", "Realm created by SDK", true)
	if err != nil {
		log.Println(err)
	}

	clientScope, err := keycloakSDK.CreateClientScope(realm.Realm, "Client_Scope_SDK", "Client Scope criado via SDK", "openid-connect")
	if err != nil {
		log.Println(err)
	}

	client, err := keycloakSDK.CreateClient(realm.Realm, "Client_API_SDK", "Client_API_SDK", "Client API criado via SDK", "openid-connect", "http://localhost:9000", clientScope.Name, true, false)
	if err != nil {
		log.Println(err)
	}

	log.Println(client.ID)
}
