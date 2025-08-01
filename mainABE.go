package main

import (
	"Obfushop/bn256"
	"Obfushop/crypto/OABE"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

func main() {
	iterations := 2000
	// 系统参数生成
	MSK, PK := OABE.Setup()

	//生成用户公私钥对
	sku, _ := rand.Int(rand.Reader, bn256.Order)
	var pku *bn256.G1
	start := time.Now() // 记录开始时间
	for i := 0; i < iterations; i++ {
		pku = new(bn256.G1).ScalarMult(PK.G1, sku)
	}
	elapsed := time.Since(start)
	fmt.Printf("Time cost (Our Additional Operation): %.6f ms\n", elapsed.Seconds()*1000/float64(iterations))

	// 敏感信息加密
	// 随机生成一个秘密
	secret, _ := rand.Int(rand.Reader, bn256.Order)
	m := new(bn256.GT).ScalarBaseMult(secret)
	//fmt.Printf("m:%v\n", m)
	// 原始策略表达式
	// policyStr := "(A AND (B OR C)) OR (D AND (t-of-(2,E,F,G)))"
	//tau := "(Age>18 AND (Man OR Student)) OR (Computer AND (t-of-(2,China,Sichuan,Teacher)))"
	//tau := "(A1 AND A2) OR (t-of-(2, A3,A4, A5))"
	//tau := "(A1 AND A2 AND A3 AND A4) OR (t-of-(2, A5,A6, A7))"
	//tau := "(A1 AND A2 AND A3 AND A4 AND A5 AND A6) OR (t-of-(2, A7, A8, A9)))"
	//tau := "(A1 AND A2 AND A3 AND A4 AND A5 AND A6 AND A7 AND A8) OR (t-of-(2, A9, A10, A11)))"
	tau := "(A1 AND A2 AND A3 AND A4 AND A5 AND A6 AND A7 AND A8 AND A9 AND A10) OR (t-of-(2, A11, A12, A13)))"
	CT, xsMap, key, sss := OABE.Encrypt(m, tau, PK)

	// 使用用户公钥生成加密的属性密钥
	//Su := map[string]bool{"Age>18": true, "Man": true}
	//"A1": true, "A2": true, "A3": true, "A4": true, "A5": true, "A6": true,"A7": true, "A8": true, "A9": true, "A10": true
	Su := map[string]bool{"A1": true, "A2": true, "A3": true, "A4": true, "A5": true, "A6": true, "A7": true, "A8": true, "A9": true, "A10": true}
	var attributeSet []string
	for key, _ := range Su {
		attributeSet = append(attributeSet, key)
	}

	SK := OABE.KeyGen(pku, MSK, PK, attributeSet)

	// 外包解密
	IR := OABE.ODecrypt(Su, CT, SK, xsMap, PK)

	// 用户解密
	var _m *bn256.GT
	start = time.Now() // 记录开始时间
	for i := 0; i < iterations; i++ {
		_m = OABE.Decrypt(IR, sku, CT)
	}
	elapsed = time.Since(start)
	fmt.Printf("Time cost (OABE Decrypt): %.6f ms\n", elapsed.Seconds()*1000/float64(iterations))

	// 验证正确性
	//fmt.Printf("_m:%v\n", _m)
	if m.String() == _m.String() {
		fmt.Println("OABE Secret recovery successful.")
	} else {
		fmt.Println("OABE Secret recovery failed.")
	}

	//BSW-ABE
	CTBSW, xsMapBSW, _, _ := OABE.Encrypt(m, tau, PK)
	var attributeSetBSW []string
	for key, _ := range Su {
		attributeSetBSW = append(attributeSetBSW, key)
	}

	BSWSK := OABE.KeyGen(PK.G1, MSK, PK, attributeSet)

	var BSW_m *bn256.GT
	start = time.Now() // 记录开始时间
	for i := 0; i < iterations; i++ {
		BSW_m = OABE.BSWDecrypt(Su, CTBSW, BSWSK, xsMapBSW, PK)
	}
	elapsed = time.Since(start)
	fmt.Printf("Time cost (BSW Decrypt): %.6f ms\n", elapsed.Seconds()*1000/float64(iterations))

	if m.String() == BSW_m.String() {
		fmt.Println("BSW-ABE Secret recovery successful.")
	} else {
		fmt.Println("BSW-ABE Secret recovery failed.")
	}

	//Ma et.al Scheme
	K2Set := make([]*bn256.G2, len(attributeSet))
	K3Set := make([]*bn256.G1, len(attributeSet))
	rSet := make([]*big.Int, len(attributeSet))
	r, _ := rand.Int(rand.Reader, bn256.Order)
	for i := 0; i < len(attributeSet); i++ {
		rSet[i], _ = rand.Int(rand.Reader, bn256.Order)
	}
	w := new(bn256.G1).ScalarBaseMult(secret)
	v := new(bn256.G1).ScalarBaseMult(big.NewInt(1232133243))
	h := new(bn256.G1).ScalarBaseMult(big.NewInt(471284298472198479))
	//KeyGen
	K0 := new(bn256.G1).Add(new(bn256.G1).ScalarBaseMult(MSK), new(bn256.G1).ScalarMult(w, r))
	K1 := new(bn256.G2).ScalarBaseMult(r)
	for i := 0; i < len(attributeSet); i++ {
		K2Set[i] = new(bn256.G2).ScalarBaseMult(rSet[i])
		u, _ := bn256.HashG1(attributeSet[i])
		temp1 := new(bn256.G1).ScalarMult(new(bn256.G1).Add(u, h), rSet[i])
		temp2 := new(bn256.G1).Neg(new(bn256.G1).ScalarMult(v, r))
		K3Set[i] = new(bn256.G1).Add(temp1, temp2)
	}
	//Additional Operation
	var theta *big.Int
	var thetaInv *big.Int
	start = time.Now() // 记录开始时间
	for i := 0; i < iterations; i++ {
		theta, _ = rand.Int(rand.Reader, bn256.Order)
		thetaInv = new(big.Int).ModInverse(theta, bn256.Order)
		K0 = new(bn256.G1).ScalarMult(K0, thetaInv)
		K1 = new(bn256.G2).ScalarMult(K1, thetaInv)
		for i := 0; i < len(attributeSet); i++ {
			K2Set[i] = new(bn256.G2).ScalarMult(K2Set[i], thetaInv)
			K3Set[i] = new(bn256.G1).ScalarMult(K3Set[i], thetaInv)
		}
	}
	elapsed = time.Since(start)
	fmt.Printf("Time cost (Ma addtional Decrypt): %.6f ms\n", elapsed.Seconds()*1000/float64(iterations))

	T := new(bn256.GT).ScalarMult(key, thetaInv)
	//Decryption
	var Ma_m *bn256.GT
	start = time.Now() // 记录开始时间
	for i := 0; i < iterations; i++ {
		Ma_m = new(bn256.GT).Add(CT.C, new(bn256.GT).Neg(new(bn256.GT).ScalarMult(T, theta)))
	}
	elapsed = time.Since(start)
	fmt.Printf("Time cost (Ma Decrypt): %.6f ms\n", elapsed.Seconds()*1000/float64(iterations))

	// 验证正确性
	if m.String() == Ma_m.String() {
		fmt.Println("Ma Secret recovery successful.")
	} else {
		fmt.Println("Ma Secret recovery failed.")
	}

	//Ge et.al Scheme
	//s, _ := rand.Int(rand.Reader, bn256.Order)
	a, _ := rand.Int(rand.Reader, bn256.Order)
	R := new(bn256.G1).Add(new(bn256.G1).ScalarBaseMult(MSK), new(bn256.G1).ScalarMult(new(bn256.G1).ScalarBaseMult(a), secret))
	L := new(bn256.G2).ScalarBaseMult(secret)
	Rx := make([]*bn256.G1, len(attributeSet))
	for i := 0; i < len(attributeSet); i++ {
		Rx[i], _ = bn256.HashG1(attributeSet[i])
		Rx[i] = new(bn256.G1).ScalarMult(Rx[i], secret)
	}
	//Addtional operation
	var _R *bn256.G1
	var _L *bn256.G2
	start = time.Now() // 记录开始时间
	for i := 0; i < iterations; i++ {
		_Rx := make([]*bn256.G1, len(attributeSet))
		_R = new(bn256.G1).ScalarMult(R, thetaInv)
		_L = new(bn256.G2).ScalarMult(L, thetaInv)
		for i := 0; i < len(attributeSet); i++ {
			_Rx[i] = new(bn256.G1).ScalarMult(Rx[i], thetaInv)
		}
	}
	elapsed = time.Since(start)
	fmt.Printf("Time cost (Ge Additional Operation): %.6f ms\n", elapsed.Seconds()*1000/float64(iterations))

	T = new(bn256.GT).ScalarMult(key, thetaInv)
	//Decryption
	var Ge_m *bn256.GT
	start = time.Now() // 记录开始时间
	for i := 0; i < iterations; i++ {
		Ge_m = new(bn256.GT).Add(CT.C, new(bn256.GT).Neg(new(bn256.GT).ScalarMult(T, theta)))
	}
	elapsed = time.Since(start)
	fmt.Printf("Time cost (Ge Decrypt): %.6f ms\n", elapsed.Seconds()*1000/float64(iterations))

	// 验证正确性
	if m.String() == Ge_m.String() {
		fmt.Println("Ge Secret recovery successful.")
	} else {
		fmt.Println("Ge Secret recovery failed.")
	}

	_R = new(bn256.G1).Neg(_R)
	_L = new(bn256.G2).Neg(_L)

	//Miao et.al Scheme
	Base := bn256.Pair(new(bn256.G1).ScalarBaseMult(MSK), new(bn256.G2).ScalarBaseMult(big.NewInt(1)))
	Miao_CT := new(bn256.GT).Add(m, new(bn256.GT).ScalarMult(Base, sss))
	Miao_CT.Add(Miao_CT, new(bn256.GT).ScalarMult(Base, r))
	Miao_C := new(bn256.G2).Add(new(bn256.G2).ScalarBaseMult(sss), new(bn256.G2).ScalarBaseMult(r))
	alpha1, _ := rand.Int(rand.Reader, bn256.Order)
	alpha2 := new(big.Int).Sub(MSK, alpha1)
	alpha2.Mod(alpha2, bn256.Order)
	atu, _ := rand.Int(rand.Reader, bn256.Order)
	SK_DU := new(bn256.G1).Add(new(bn256.G1).ScalarBaseMult(alpha2), new(bn256.G1).ScalarBaseMult(atu))
	phi_DU := bn256.Pair(new(bn256.G1).ScalarBaseMult(sss), new(bn256.G2).ScalarBaseMult(alpha2))
	phi_DU.Add(phi_DU, bn256.Pair(new(bn256.G1).ScalarBaseMult(r), new(bn256.G2).ScalarBaseMult(alpha2)))
	phi_DU.Add(phi_DU, new(bn256.GT).Neg(bn256.Pair(new(bn256.G1).ScalarBaseMult(atu), new(bn256.G2).ScalarBaseMult(sss))))
	phi_DU.Add(phi_DU, new(bn256.GT).Neg(bn256.Pair(new(bn256.G1).ScalarBaseMult(atu), new(bn256.G2).ScalarBaseMult(r))))
	var Miao_m *bn256.GT
	start = time.Now() // 记录开始时间
	for i := 0; i < iterations; i++ {
		Miao_m = new(bn256.GT).Add(Miao_CT, new(bn256.GT).Neg(phi_DU))
		Miao_m.Add(Miao_m, new(bn256.GT).Neg(bn256.Pair(SK_DU, Miao_C)))
	}
	elapsed = time.Since(start)
	fmt.Printf("Time cost (Miao Decrypt): %.6f ms\n", elapsed.Seconds()*1000/float64(iterations))

	// 验证正确性
	if m.String() == Miao_m.String() {
		fmt.Println("Miao Secret recovery successful.")
	} else {
		fmt.Println("Miao Secret recovery failed.")
	}

}

//Test Result (attr=2,4,6,8,10)
// Our Decrypt:1.108641 ms;1.029680 ms;1.015491 ms; 1.035348 ms;1.002222 ms
// Our Addition:0.089075 ms;0.084385 ms;0.082842 ms;0.082639 ms;0.083110 ms

//BSW Decrypt:9.374658 ms;15.847881 ms;22.257073 ms;29.000092 ms;35.782498 ms

//Ma Decrypt:1.047784 ms;1.017488 ms;1.012373 ms;1.041507 ms;1.057265 ms
//Ma Addition:1.426772 ms;2.419020 ms;3.322068 ms;4.289441 ms;5.227823 ms

//Ge Decrypt:1.061963 ms;1.014540 ms;1.012828 ms;1.036949 ms;1.037816 ms
//Ge Addition:0.607187 ms;0.764293 ms;0.905052 ms;1.066154 ms; 1.386781 ms

//Miao Decrypt:1.368643 ms;1.379879 ms;1.349876 ms;1.338974 ms;1.473916 ms
