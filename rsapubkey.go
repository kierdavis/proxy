package proxy

import (
	"crypto/rsa"
	"encoding/asn1"
	"fmt"
)

var rsaAlgorithm = asn1.ObjectIdentifier{1,2,840,113549,1,1,1}

func oidsEqual(a, b asn1.ObjectIdentifier) (ok bool) {
	if len(a) != len(b) {
		return false
	}
	
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	
	return true
}

type publicKeyStructure struct {
	Algorithm publicKeyAlgorithm
	SubjectPublicKey asn1.BitString
}

type publicKeyAlgorithm struct {
	Algorithm asn1.ObjectIdentifier
	Parameters interface{} `optional`
}

func decodePublicKey(data []byte) (key rsa.PublicKey, err error) {
	var pks publicKeyStructure
	_, err = asn1.Unmarshal(data, &pks)
	if err != nil {
		return key, err
	}
	
	if !oidsEqual(pks.Algorithm.Algorithm, rsaAlgorithm) {
		return key, fmt.Errorf("Unexpected public key algorithm (object identifier %v)", pks.Algorithm.Algorithm)
	}
	
	_, err = asn1.Unmarshal(pks.SubjectPublicKey.Bytes, &key)
	return key, err
}

func encodePublicKey(key rsa.PublicKey) (data []byte, err error) {
	subjectPublicKey, err := asn1.Marshal(key)
	if err != nil {
		return nil, err
	}
	
	pks := publicKeyStructure{
		Algorithm: publicKeyAlgorithm{
			Algorithm: rsaAlgorithm,
			Parameters: "foo",
		},
		SubjectPublicKey: asn1.BitString{subjectPublicKey, len(subjectPublicKey)*8},
	}
	
	return asn1.Marshal(pks)
}
