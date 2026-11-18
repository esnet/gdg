package api

import "log"

func (s *DashNGoImpl) EncodeValue(in string) string {
	newVal, err := s.Encoder.EncodeValue(in)
	if err != nil {
		log.Fatal(err)
	}

	return newVal
}

func (s *DashNGoImpl) DecodeValue(in string) string {
	newVal, err := s.Encoder.DecodeValue(in)
	if err != nil {
		log.Fatal(err)
	}

	return newVal
}
