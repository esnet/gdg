package service

// IsEnterprise will return a valid response if the grafana version is running an enterprise version
func (s *DashNGoImpl) IsEnterprise() bool {
	r, err := s.GetClient().Licensing.GetStatus()
	if err != nil {
		return false
	}

	return r.IsSuccess()
}
