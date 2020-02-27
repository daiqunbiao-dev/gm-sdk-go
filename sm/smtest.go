package sm

import (
	"crypto/rand"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"github.com/tjfoc/gmsm/sm2"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

type Sm2Test struct {

}

func(Sm2Test *Sm2Test) VerifySm2(Sm2FilePath string){
	cert, err := sm2.ReadCertificateFromPem(Sm2FilePath)
	if err != nil {
		fmt.Printf("failed to read cert file")
	}
	err = cert.CheckSignature(cert.SignatureAlgorithm, cert.RawTBSCertificate, cert.Signature)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("CheckSignature ok\n")
	}
}

func(smtest *Sm2Test) TestSm2() {
	var (
		oidExtensionSubjectKeyId          = []int{2, 5, 29, 14}
	)
	priv, err := sm2.GenerateKey() // 生成密钥对
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v\n", priv.Curve.IsOnCurve(priv.X, priv.Y)) // 验证是否为sm2的曲线
	pub := &priv.PublicKey
	msg := []byte("123456")
	d0, err := pub.Encrypt(msg)
	if err != nil {
		fmt.Printf("Error: failed to encrypt %s: %v\n", msg, err)
		return
	}
	fmt.Printf("Cipher text = %v\n", d0)
	d1, err := priv.Decrypt(d0)
	if err != nil {
		fmt.Printf("Error: failed to decrypt: %v\n", err)
	}
	fmt.Printf("clear text = %s\n", d1)
	ok, err := sm2.WritePrivateKeytoPem("priv.pem", priv, nil) // 生成密钥文件
	if ok != true {
		log.Fatal(err)
	}
	pubKey, _ := priv.Public().(*sm2.PublicKey)
	ok, err = sm2.WritePublicKeytoPem("pub.pem", pubKey, nil) // 生成公钥文件
	if ok != true {
		log.Fatal(err)
	}
	msg = []byte("test")
	err = ioutil.WriteFile("ifile", msg, os.FileMode(0644)) // 生成测试文件
	if err != nil {
		log.Fatal(err)
	}
	privKey, err := sm2.ReadPrivateKeyFromPem("priv.pem", nil) // 读取密钥
	if err != nil {
		log.Fatal(err)
	}
	pubKey, err = sm2.ReadPublicKeyFromPem("pub.pem", nil) // 读取公钥
	if err != nil {
		log.Fatal(err)
	}
	msg, _ = ioutil.ReadFile("ifile")                // 从文件读取数据
	sign, err := privKey.Sign(rand.Reader, msg, nil) // 签名
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("ofile", sign, os.FileMode(0644))
	if err != nil {
		log.Fatal(err)
	}
	signdata, _ := ioutil.ReadFile("ofile")
	ok = privKey.Verify(msg, signdata) // 密钥验证
	if ok != true {
		fmt.Printf("Verify error\n")
	} else {
		fmt.Printf("Verify ok\n")
	}
	ok = pubKey.Verify(msg, signdata) // 公钥验证
	if ok != true {
		fmt.Printf("Verify error\n")
	} else {
		fmt.Printf("Verify ok\n")
	}
	templateReq := sm2.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   "test.example.com",
			Organization: []string{"Test"},
		},
		//		SignatureAlgorithm: ECDSAWithSHA256,
		SignatureAlgorithm: sm2.SM2WithSM3,
	}
	_, err = sm2.CreateCertificateRequestToPem("req.pem", &templateReq, privKey)
	if err != nil {
		log.Fatal(err)
	}
	req, err := sm2.ReadCertificateRequestFromPem("req.pem")
	if err != nil {
		log.Fatal(err)
	}
	err = req.CheckSignature()
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("CheckSignature ok\n")
	}
	testExtKeyUsage := []sm2.ExtKeyUsage{sm2.ExtKeyUsageClientAuth, sm2.ExtKeyUsageServerAuth}
	testUnknownExtKeyUsage := []asn1.ObjectIdentifier{[]int{1, 2, 3}, []int{2, 59, 1}}
	extraExtensionData := []byte("extra extension")
	commonName := "test.example.com"
	template := sm2.Certificate{
		// SerialNumber is negative to ensure that negative
		// values are parsed. This is due to the prevalence of
		// buggy code that produces certificates with negative
		// serial numbers.
		SerialNumber: big.NewInt(-1),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"TEST"},
			Country:      []string{"China"},
			ExtraNames: []pkix.AttributeTypeAndValue{
				{
					Type:  []int{2, 5, 4, 42},
					Value: "Gopher",
				},
				// This should override the Country, above.
				{
					Type:  []int{2, 5, 4, 6},
					Value: "NL",
				},
			},
		},
		NotBefore: time.Unix(1000, 0),
		NotAfter:  time.Unix(100000, 0),

		//		SignatureAlgorithm: ECDSAWithSHA256,
		SignatureAlgorithm: sm2.SM2WithSM3,

		SubjectKeyId: []byte{1, 2, 3, 4},
		KeyUsage:     sm2.KeyUsageCertSign,

		ExtKeyUsage:        testExtKeyUsage,
		UnknownExtKeyUsage: testUnknownExtKeyUsage,

		BasicConstraintsValid: true,
		IsCA: true,

		OCSPServer:            []string{"http://ocsp.example.com"},
		IssuingCertificateURL: []string{"http://crt.example.com/ca1.crt"},

		DNSNames:       []string{"test.example.com"},
		EmailAddresses: []string{"gopher@golang.org"},
		IPAddresses:    []net.IP{net.IPv4(127, 0, 0, 1).To4(), net.ParseIP("2001:4860:0:2001::68")},

		PolicyIdentifiers:   []asn1.ObjectIdentifier{[]int{1, 2, 3}},
		PermittedDNSDomains: []string{".example.com", "example.com"},

		CRLDistributionPoints: []string{"http://crl1.example.com/ca1.crl", "http://crl2.example.com/ca1.crl"},

		ExtraExtensions: []pkix.Extension{
			{
				Id:    []int{1, 2, 3, 4},
				Value: extraExtensionData,
			},
			// This extension should override the SubjectKeyId, above.
			{
				Id:       oidExtensionSubjectKeyId,
				Critical: false,
				Value:    []byte{0x04, 0x04, 4, 3, 2, 1},
			},
		},
	}
	pubKey, _ = priv.Public().(*sm2.PublicKey)
	ok, _ = sm2.CreateCertificateToPem("cert.pem", &template, &template, pubKey, privKey)
	if ok != true {
		fmt.Printf("failed to create cert file\n")
	}
	cert, err := sm2.ReadCertificateFromPem("cert.pem")
	if err != nil {
		fmt.Printf("failed to read cert file")
	}
	err = cert.CheckSignature(cert.SignatureAlgorithm, cert.RawTBSCertificate, cert.Signature)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("CheckSignature ok\n")
	}
}
