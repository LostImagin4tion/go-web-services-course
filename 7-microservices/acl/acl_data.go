package acl

import (
	"encoding/json"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"slices"
	"stepikGoWebServices/generated/service"
	"stepikGoWebServices/services"
	"strings"
)

type Data struct {
	Map map[string]map[string][]string
}

func NewAclData(aclRaw string) (*Data, error) {
	var rawMap = make(map[string][]string)

	var err = json.Unmarshal([]byte(aclRaw), &rawMap)
	if err != nil {
		return nil, err
	}

	var convertedMap = make(map[string]map[string][]string)

	for consumer, allowedMethods := range rawMap {
		convertedMap[consumer] = make(map[string][]string)
		var consumerMethods = convertedMap[consumer]

		for _, allowedMethod := range allowedMethods {
			var split = strings.Split(allowedMethod, "/")
			var serviceName = split[1]
			var method = split[2]

			switch serviceName {
			case service.Admin_ServiceDesc.ServiceName:
				if method == "*" {
					consumerMethods[serviceName] = services.AdminServiceMethods
				} else if slices.Contains(services.AdminServiceMethods, allowedMethod) {
					consumerMethods[serviceName] = append(
						consumerMethods[serviceName],
						allowedMethod,
					)
				} else {
					log.Printf("Unknown method %s for service %s", allowedMethod, serviceName)
				}

			case service.BusinessLogic_ServiceDesc.ServiceName:
				if method == "*" {
					consumerMethods[serviceName] = services.BusinessLogicMethods
				} else if slices.Contains(services.BusinessLogicMethods, allowedMethod) {
					consumerMethods[serviceName] = append(
						consumerMethods[serviceName],
						allowedMethod,
					)
				} else {
					log.Printf("Unknown method %s for service %s", allowedMethod, serviceName)
				}

			default:
				log.Println("Unknown service ", serviceName)
				continue
			}
		}
	}

	return &Data{Map: convertedMap}, nil
}

func (a *Data) ValidateAcl(
	consumer string,
	fullMethod string,
) error {
	var accessibleClasses, exists = a.Map[consumer]

	if !exists {
		return status.Error(codes.Unauthenticated, "unknown consumer")
	}

	var split = strings.Split(fullMethod, "/")
	var serviceName = split[1]

	//fmt.Printf("Acl data %v\n", acl)
	//fmt.Printf("Consumer %v Service %s Method %v\n\n", consumer, serviceName, info.FullMethod)

	accessibleMethods, exists := accessibleClasses[serviceName]

	if !exists {
		return status.Error(codes.Unauthenticated, "forbidden")
	}

	if !slices.Contains(accessibleMethods, fullMethod) {
		return status.Error(codes.Unauthenticated, "access forbidden")
	}

	return nil
}
