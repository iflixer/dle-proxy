package server

import "strings"

func (s *Service) traefikConfig() string {
	tpl := `
	[providers.http.routers]\n
	[providers.http.routers.[host_]]\n
    	rule = Host([host])
    	service = [host_]\n`

	domains, _ := s.domainService.GetDomains()
	result := "[providers.http]\n"
	for _, d := range domains {
		row := tpl
		row = strings.ReplaceAll(row, "[host]", d.HostPublic)
		host_ := strings.ReplaceAll(d.HostPublic, ".", "_")
		row = strings.ReplaceAll(row, "[host_]", host_)
		result = result + row
	}
	return result
}
