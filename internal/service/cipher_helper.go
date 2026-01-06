package service

import "log"

func (s *DashNGoImpl) EncodeValue(in string) string {
	newVal, err := s.encoder.EncodeValue(in)
	if err != nil {
		log.Fatal(err)
	}

	return newVal
}

func (s *DashNGoImpl) DecodeValue(in string) string {
	newVal, err := s.encoder.DecodeValue(in)
	if err != nil {
		log.Fatal(err)
	}

	return newVal
}
