package server

import "strings"

func (s *Service) traefikConfig() string {
	tpl := "" +
		"  [host_]:\n" +
		"   rule: Host(`[host]`)\n" +
		"   service: cis-proxy@docker\n"

	domains, _ := s.domainService.GetDomains()
	result := "http:\n" +
		" routers:\n"
	for _, d := range domains {
		row := tpl
		row = strings.ReplaceAll(row, "[host]", d.HostPublic)
		host_ := strings.ReplaceAll(d.HostPublic, ".", "_")
		row = strings.ReplaceAll(row, "[host_]", host_)
		result = result + row
	}
	return result
}
