package money

import (
	"hara.sh/alloy/money/currency"
)

// defaultManager is the default singleton instance of Manager used by the factory methods.
var defaultManager = NewManager()

// FromUSD creates a new Value instance with US Dollars
func FromUSD(amount int64) *Value {
	return defaultManager.Create(amount, currency.USD)
}

// FromEUR creates a new Value instance with Euros
func FromEUR(amount int64) *Value {
	return defaultManager.Create(amount, currency.EUR)
}

// FromGBP creates a new Value instance with British Pounds
func FromGBP(amount int64) *Value {
	return defaultManager.Create(amount, currency.GBP)
}

// FromJPY creates a new Value instance with Japanese Yen
func FromJPY(amount int64) *Value {
	return defaultManager.Create(amount, currency.JPY)
}

// FromCHF creates a new Value instance with Swiss Francs
func FromCHF(amount int64) *Value {
	return defaultManager.Create(amount, currency.CHF)
}

// FromCAD creates a new Value instance with Canadian Dollars
func FromCAD(amount int64) *Value {
	return defaultManager.Create(amount, currency.CAD)
}

// FromAUD creates a new Value instance with Australian Dollars
func FromAUD(amount int64) *Value {
	return defaultManager.Create(amount, currency.AUD)
}

// FromCNY creates a new Value instance with Chinese Yuan
func FromCNY(amount int64) *Value {
	return defaultManager.Create(amount, currency.CNY)
}

// FromINR creates a new Value instance with Indian Rupees
func FromINR(amount int64) *Value {
	return defaultManager.Create(amount, currency.INR)
}

// FromBRL creates a new Value instance with Brazilian Reais
func FromBRL(amount int64) *Value {
	return defaultManager.Create(amount, currency.BRL)
}

// FromMXN creates a new Value instance with Mexican Pesos
func FromMXN(amount int64) *Value {
	return defaultManager.Create(amount, currency.MXN)
}

// FromNZD creates a new Value instance with New Zealand Dollars
func FromNZD(amount int64) *Value {
	return defaultManager.Create(amount, currency.NZD)
}

// FromSGD creates a new Value instance with Singapore Dollars
func FromSGD(amount int64) *Value {
	return defaultManager.Create(amount, currency.SGD)
}

// FromHKD creates a new Value instance with Hong Kong Dollars
func FromHKD(amount int64) *Value {
	return defaultManager.Create(amount, currency.HKD)
}

// FromNOK creates a new Value instance with Norwegian Kroner
func FromNOK(amount int64) *Value {
	return defaultManager.Create(amount, currency.NOK)
}

// FromSEK creates a new Value instance with Swedish Kronor
func FromSEK(amount int64) *Value {
	return defaultManager.Create(amount, currency.SEK)
}

// FromDKK creates a new Value instance with Danish Kroner
func FromDKK(amount int64) *Value {
	return defaultManager.Create(amount, currency.DKK)
}

// FromRUB creates a new Value instance with Russian Rubles
func FromRUB(amount int64) *Value {
	return defaultManager.Create(amount, currency.RUB)
}

// FromZAR creates a new Value instance with South African Rand
func FromZAR(amount int64) *Value {
	return defaultManager.Create(amount, currency.ZAR)
}

// FromTRY creates a new Value instance with Turkish Lira
func FromTRY(amount int64) *Value {
	return defaultManager.Create(amount, currency.TRY)
}

// FromKRW creates a new Value instance with South Korean Won
func FromKRW(amount int64) *Value {
	return defaultManager.Create(amount, currency.KRW)
}

// FromPLN creates a new Value instance with Polish Zloty
func FromPLN(amount int64) *Value {
	return defaultManager.Create(amount, currency.PLN)
}

// FromTHB creates a new Value instance with Thai Baht
func FromTHB(amount int64) *Value {
	return defaultManager.Create(amount, currency.THB)
}

// FromIDR creates a new Value instance with Indonesian Rupiah
func FromIDR(amount int64) *Value {
	return defaultManager.Create(amount, currency.IDR)
}

// FromMYR creates a new Value instance with Malaysian Ringgit
func FromMYR(amount int64) *Value {
	return defaultManager.Create(amount, currency.MYR)
}

// FromPHP creates a new Value instance with Philippine Pesos
func FromPHP(amount int64) *Value {
	return defaultManager.Create(amount, currency.PHP)
}

// FromCZK creates a new Value instance with Czech Koruny
func FromCZK(amount int64) *Value {
	return defaultManager.Create(amount, currency.CZK)
}

// FromHUF creates a new Value instance with Hungarian Forints
func FromHUF(amount int64) *Value {
	return defaultManager.Create(amount, currency.HUF)
}

// FromILS creates a new Value instance with Israeli NewExchange Shekels
func FromILS(amount int64) *Value {
	return defaultManager.Create(amount, currency.ILS)
}

// FromCLP creates a new Value instance with Chilean Pesos
func FromCLP(amount int64) *Value {
	return defaultManager.Create(amount, currency.CLP)
}

// FromAED creates a new Value instance with the United Arab Emirates Dirhams
func FromAED(amount int64) *Value {
	return defaultManager.Create(amount, currency.AED)
}

// FromSAR creates a new Value instance with Saudi Riyals
func FromSAR(amount int64) *Value {
	return defaultManager.Create(amount, currency.SAR)
}

// FromARS creates a new Value instance with Argentine Pesos
func FromARS(amount int64) *Value {
	return defaultManager.Create(amount, currency.ARS)
}

// FromCOP creates a new Value instance with Colombian Pesos
func FromCOP(amount int64) *Value {
	return defaultManager.Create(amount, currency.COP)
}

// FromTWD creates a new Value instance with NewMoney Taiwan Dollars
func FromTWD(amount int64) *Value {
	return defaultManager.Create(amount, currency.TWD)
}

// FromVND creates a new Value instance with Vietnamese Dong
func FromVND(amount int64) *Value {
	return defaultManager.Create(amount, currency.VND)
}

// FromAFN creates a new Value instance with AFN currency
func FromAFN(amount int64) *Value {
	return defaultManager.Create(amount, currency.AFN)
}

// FromALL creates a new Value instance with ALL currency
func FromALL(amount int64) *Value {
	return defaultManager.Create(amount, currency.ALL)
}

// FromAMD creates a new Value instance with AMD currency
func FromAMD(amount int64) *Value {
	return defaultManager.Create(amount, currency.AMD)
}

// FromANG creates a new Value instance with ANG currency
func FromANG(amount int64) *Value {
	return defaultManager.Create(amount, currency.ANG)
}

// FromAOA creates a new Value instance with AOA currency
func FromAOA(amount int64) *Value {
	return defaultManager.Create(amount, currency.AOA)
}

// FromAWG creates a new Value instance with AWG currency
func FromAWG(amount int64) *Value {
	return defaultManager.Create(amount, currency.AWG)
}

// FromAZN creates a new Value instance with AZN currency
func FromAZN(amount int64) *Value {
	return defaultManager.Create(amount, currency.AZN)
}

// FromBAM creates a new Value instance with BAM currency
func FromBAM(amount int64) *Value {
	return defaultManager.Create(amount, currency.BAM)
}

// FromBBD creates a new Value instance with BBD currency
func FromBBD(amount int64) *Value {
	return defaultManager.Create(amount, currency.BBD)
}

// FromBDT creates a new Value instance with BDT currency
func FromBDT(amount int64) *Value {
	return defaultManager.Create(amount, currency.BDT)
}

// FromBGN creates a new Value instance with BGN currency
func FromBGN(amount int64) *Value {
	return defaultManager.Create(amount, currency.BGN)
}

// FromBHD creates a new Value instance with BHD currency
func FromBHD(amount int64) *Value {
	return defaultManager.Create(amount, currency.BHD)
}

// FromBIF creates a new Value instance with BIF currency
func FromBIF(amount int64) *Value {
	return defaultManager.Create(amount, currency.BIF)
}

// FromBMD creates a new Value instance with BMD currency
func FromBMD(amount int64) *Value {
	return defaultManager.Create(amount, currency.BMD)
}

// FromBND creates a new Value instance with BND currency
func FromBND(amount int64) *Value {
	return defaultManager.Create(amount, currency.BND)
}

// FromBOB creates a new Value instance with BOB currency
func FromBOB(amount int64) *Value {
	return defaultManager.Create(amount, currency.BOB)
}

// FromBOV creates a new Value instance with BOV currency
func FromBOV(amount int64) *Value {
	return defaultManager.Create(amount, currency.BOV)
}

// FromBSD creates a new Value instance with BSD currency
func FromBSD(amount int64) *Value {
	return defaultManager.Create(amount, currency.BSD)
}

// FromBTN creates a new Value instance with BTN currency
func FromBTN(amount int64) *Value {
	return defaultManager.Create(amount, currency.BTN)
}

// FromBWP creates a new Value instance with BWP currency
func FromBWP(amount int64) *Value {
	return defaultManager.Create(amount, currency.BWP)
}

// FromBYN creates a new Value instance with BYN currency
func FromBYN(amount int64) *Value {
	return defaultManager.Create(amount, currency.BYN)
}

// FromBZD creates a new Value instance with BZD currency
func FromBZD(amount int64) *Value {
	return defaultManager.Create(amount, currency.BZD)
}

// FromCDF creates a new Value instance with CDF currency
func FromCDF(amount int64) *Value {
	return defaultManager.Create(amount, currency.CDF)
}

// FromCHE creates a new Value instance with CHE currency
func FromCHE(amount int64) *Value {
	return defaultManager.Create(amount, currency.CHE)
}

// FromCHW creates a new Value instance with CHW currency
func FromCHW(amount int64) *Value {
	return defaultManager.Create(amount, currency.CHW)
}

// FromCLF creates a new Value instance with CLF currency
func FromCLF(amount int64) *Value {
	return defaultManager.Create(amount, currency.CLF)
}

// FromCOU creates a new Value instance with COU currency
func FromCOU(amount int64) *Value {
	return defaultManager.Create(amount, currency.COU)
}

// FromCRC creates a new Value instance with CRC currency
func FromCRC(amount int64) *Value {
	return defaultManager.Create(amount, currency.CRC)
}

// FromCUC creates a new Value instance with CUC currency
func FromCUC(amount int64) *Value {
	return defaultManager.Create(amount, currency.CUC)
}

// FromCUP creates a new Value instance with CUP currency
func FromCUP(amount int64) *Value {
	return defaultManager.Create(amount, currency.CUP)
}

// FromCVE creates a new Value instance with CVE currency
func FromCVE(amount int64) *Value {
	return defaultManager.Create(amount, currency.CVE)
}

// FromDJF creates a new Value instance with DJF currency
func FromDJF(amount int64) *Value {
	return defaultManager.Create(amount, currency.DJF)
}

// FromDOP creates a new Value instance with DOP currency
func FromDOP(amount int64) *Value {
	return defaultManager.Create(amount, currency.DOP)
}

// FromDZD creates a new Value instance with DZD currency
func FromDZD(amount int64) *Value {
	return defaultManager.Create(amount, currency.DZD)
}

// FromEGP creates a new Value instance with EGP currency
func FromEGP(amount int64) *Value {
	return defaultManager.Create(amount, currency.EGP)
}

// FromERN creates a new Value instance with ERN currency
func FromERN(amount int64) *Value {
	return defaultManager.Create(amount, currency.ERN)
}

// FromETB creates a new Value instance with ETB currency
func FromETB(amount int64) *Value {
	return defaultManager.Create(amount, currency.ETB)
}

// FromFJD creates a new Value instance with FJD currency
func FromFJD(amount int64) *Value {
	return defaultManager.Create(amount, currency.FJD)
}

// FromFKP creates a new Value instance with FKP currency
func FromFKP(amount int64) *Value {
	return defaultManager.Create(amount, currency.FKP)
}

// FromGEL creates a new Value instance with GEL currency
func FromGEL(amount int64) *Value {
	return defaultManager.Create(amount, currency.GEL)
}

// FromGHS creates a new Value instance with GHS currency
func FromGHS(amount int64) *Value {
	return defaultManager.Create(amount, currency.GHS)
}

// FromGIP creates a new Value instance with GIP currency
func FromGIP(amount int64) *Value {
	return defaultManager.Create(amount, currency.GIP)
}

// FromGMD creates a new Value instance with GMD currency
func FromGMD(amount int64) *Value {
	return defaultManager.Create(amount, currency.GMD)
}

// FromGNF creates a new Value instance with GNF currency
func FromGNF(amount int64) *Value {
	return defaultManager.Create(amount, currency.GNF)
}

// FromGTQ creates a new Value instance with GTQ currency
func FromGTQ(amount int64) *Value {
	return defaultManager.Create(amount, currency.GTQ)
}

// FromGYD creates a new Value instance with GYD currency
func FromGYD(amount int64) *Value {
	return defaultManager.Create(amount, currency.GYD)
}

// FromHNL creates a new Value instance with HNL currency
func FromHNL(amount int64) *Value {
	return defaultManager.Create(amount, currency.HNL)
}

// FromHTG creates a new Value instance with HTG currency
func FromHTG(amount int64) *Value {
	return defaultManager.Create(amount, currency.HTG)
}

// FromIQD creates a new Value instance with IQD currency
func FromIQD(amount int64) *Value {
	return defaultManager.Create(amount, currency.IQD)
}

// FromIRR creates a new Value instance with IRR currency
func FromIRR(amount int64) *Value {
	return defaultManager.Create(amount, currency.IRR)
}

// FromISK creates a new Value instance with ISK currency
func FromISK(amount int64) *Value {
	return defaultManager.Create(amount, currency.ISK)
}

// FromJMD creates a new Value instance with JMD currency
func FromJMD(amount int64) *Value {
	return defaultManager.Create(amount, currency.JMD)
}

// FromJOD creates a new Value instance with JOD currency
func FromJOD(amount int64) *Value {
	return defaultManager.Create(amount, currency.JOD)
}

// FromKES creates a new Value instance with KES currency
func FromKES(amount int64) *Value {
	return defaultManager.Create(amount, currency.KES)
}

// FromKGS creates a new Value instance with KGS currency
func FromKGS(amount int64) *Value {
	return defaultManager.Create(amount, currency.KGS)
}

// FromKHR creates a new Value instance with KHR currency
func FromKHR(amount int64) *Value {
	return defaultManager.Create(amount, currency.KHR)
}

// FromKMF creates a new Value instance with KMF currency
func FromKMF(amount int64) *Value {
	return defaultManager.Create(amount, currency.KMF)
}

// FromKPW creates a new Value instance with KPW currency
func FromKPW(amount int64) *Value {
	return defaultManager.Create(amount, currency.KPW)
}

// FromKWD creates a new Value instance with KWD currency
func FromKWD(amount int64) *Value {
	return defaultManager.Create(amount, currency.KWD)
}

// FromKYD creates a new Value instance with KYD currency
func FromKYD(amount int64) *Value {
	return defaultManager.Create(amount, currency.KYD)
}

// FromKZT creates a new Value instance with KZT currency
func FromKZT(amount int64) *Value {
	return defaultManager.Create(amount, currency.KZT)
}

// FromLAK creates a new Value instance with LAK currency
func FromLAK(amount int64) *Value {
	return defaultManager.Create(amount, currency.LAK)
}

// FromLBP creates a new Value instance with LBP currency
func FromLBP(amount int64) *Value {
	return defaultManager.Create(amount, currency.LBP)
}

// FromLKR creates a new Value instance with LKR currency
func FromLKR(amount int64) *Value {
	return defaultManager.Create(amount, currency.LKR)
}

// FromLRD creates a new Value instance with LRD currency
func FromLRD(amount int64) *Value {
	return defaultManager.Create(amount, currency.LRD)
}

// FromLSL creates a new Value instance with LSL currency
func FromLSL(amount int64) *Value {
	return defaultManager.Create(amount, currency.LSL)
}

// FromLYD creates a new Value instance with LYD currency
func FromLYD(amount int64) *Value {
	return defaultManager.Create(amount, currency.LYD)
}

// FromMAD creates a new Value instance with MAD currency
func FromMAD(amount int64) *Value {
	return defaultManager.Create(amount, currency.MAD)
}

// FromMDL creates a new Value instance with MDL currency
func FromMDL(amount int64) *Value {
	return defaultManager.Create(amount, currency.MDL)
}

// FromMGA creates a new Value instance with MGA currency
func FromMGA(amount int64) *Value {
	return defaultManager.Create(amount, currency.MGA)
}

// FromMKD creates a new Value instance with MKD currency
func FromMKD(amount int64) *Value {
	return defaultManager.Create(amount, currency.MKD)
}

// FromMMK creates a new Value instance with MMK currency
func FromMMK(amount int64) *Value {
	return defaultManager.Create(amount, currency.MMK)
}

// FromMNT creates a new Value instance with MNT currency
func FromMNT(amount int64) *Value {
	return defaultManager.Create(amount, currency.MNT)
}

// FromMOP creates a new Value instance with MOP currency
func FromMOP(amount int64) *Value {
	return defaultManager.Create(amount, currency.MOP)
}

// FromMUR creates a new Value instance with MUR currency
func FromMUR(amount int64) *Value {
	return defaultManager.Create(amount, currency.MUR)
}

// FromMRU creates a new Value instance with MRU currency
func FromMRU(amount int64) *Value {
	return defaultManager.Create(amount, currency.MRU)
}

// FromMVR creates a new Value instance with MVR currency
func FromMVR(amount int64) *Value {
	return defaultManager.Create(amount, currency.MVR)
}

// FromMWK creates a new Value instance with MWK currency
func FromMWK(amount int64) *Value {
	return defaultManager.Create(amount, currency.MWK)
}

// FromMXV creates a new Value instance with MXV currency
func FromMXV(amount int64) *Value {
	return defaultManager.Create(amount, currency.MXV)
}

// FromMZN creates a new Value instance with MZN currency
func FromMZN(amount int64) *Value {
	return defaultManager.Create(amount, currency.MZN)
}

// FromNAD creates a new Value instance with NAD currency
func FromNAD(amount int64) *Value {
	return defaultManager.Create(amount, currency.NAD)
}

// FromNGN creates a new Value instance with NGN currency
func FromNGN(amount int64) *Value {
	return defaultManager.Create(amount, currency.NGN)
}

// FromNIO creates a new Value instance with NIO currency
func FromNIO(amount int64) *Value {
	return defaultManager.Create(amount, currency.NIO)
}

// FromNPR creates a new Value instance with NPR currency
func FromNPR(amount int64) *Value {
	return defaultManager.Create(amount, currency.NPR)
}

// FromOMR creates a new Value instance with OMR currency
func FromOMR(amount int64) *Value {
	return defaultManager.Create(amount, currency.OMR)
}

// FromPAB creates a new Value instance with PAB currency
func FromPAB(amount int64) *Value {
	return defaultManager.Create(amount, currency.PAB)
}

// FromPEN creates a new Value instance with PEN currency
func FromPEN(amount int64) *Value {
	return defaultManager.Create(amount, currency.PEN)
}

// FromPGK creates a new Value instance with PGK currency
func FromPGK(amount int64) *Value {
	return defaultManager.Create(amount, currency.PGK)
}

// FromPKR creates a new Value instance with PKR currency
func FromPKR(amount int64) *Value {
	return defaultManager.Create(amount, currency.PKR)
}

// FromPYG creates a new Value instance with PYG currency
func FromPYG(amount int64) *Value {
	return defaultManager.Create(amount, currency.PYG)
}

// FromQAR creates a new Value instance with QAR currency
func FromQAR(amount int64) *Value {
	return defaultManager.Create(amount, currency.QAR)
}

// FromRON creates a new Value instance with RON currency
func FromRON(amount int64) *Value {
	return defaultManager.Create(amount, currency.RON)
}

// FromRSD creates a new Value instance with RSD currency
func FromRSD(amount int64) *Value {
	return defaultManager.Create(amount, currency.RSD)
}

// FromRWF creates a new Value instance with RWF currency
func FromRWF(amount int64) *Value {
	return defaultManager.Create(amount, currency.RWF)
}

// FromSBD creates a new Value instance with SBD currency
func FromSBD(amount int64) *Value {
	return defaultManager.Create(amount, currency.SBD)
}

// FromSCR creates a new Value instance with SCR currency
func FromSCR(amount int64) *Value {
	return defaultManager.Create(amount, currency.SCR)
}

// FromSDG creates a new Value instance with SDG currency
func FromSDG(amount int64) *Value {
	return defaultManager.Create(amount, currency.SDG)
}

// FromSHP creates a new Value instance with SHP currency
func FromSHP(amount int64) *Value {
	return defaultManager.Create(amount, currency.SHP)
}

// FromSLE creates a new Value instance with SLE currency
func FromSLE(amount int64) *Value {
	return defaultManager.Create(amount, currency.SLE)
}

// FromSLL creates a new Value instance with SLL currency
func FromSLL(amount int64) *Value {
	return defaultManager.Create(amount, currency.SLL)
}

// FromSOS creates a new Value instance with SOS currency
func FromSOS(amount int64) *Value {
	return defaultManager.Create(amount, currency.SOS)
}

// FromSRD creates a new Value instance with SRD currency
func FromSRD(amount int64) *Value {
	return defaultManager.Create(amount, currency.SRD)
}

// FromSSP creates a new Value instance with SSP currency
func FromSSP(amount int64) *Value {
	return defaultManager.Create(amount, currency.SSP)
}

// FromSTN creates a new Value instance with STN currency
func FromSTN(amount int64) *Value {
	return defaultManager.Create(amount, currency.STN)
}

// FromSVC creates a new Value instance with SVC currency
func FromSVC(amount int64) *Value {
	return defaultManager.Create(amount, currency.SVC)
}

// FromSYP creates a new Value instance with SYP currency
func FromSYP(amount int64) *Value {
	return defaultManager.Create(amount, currency.SYP)
}

// FromSZL creates a new Value instance with SZL currency
func FromSZL(amount int64) *Value {
	return defaultManager.Create(amount, currency.SZL)
}

// FromTJS creates a new Value instance with TJS currency
func FromTJS(amount int64) *Value {
	return defaultManager.Create(amount, currency.TJS)
}

// FromTMT creates a new Value instance with TMT currency
func FromTMT(amount int64) *Value {
	return defaultManager.Create(amount, currency.TMT)
}

// FromTND creates a new Value instance with TND currency
func FromTND(amount int64) *Value {
	return defaultManager.Create(amount, currency.TND)
}

// FromTOP creates a new Value instance with TOP currency
func FromTOP(amount int64) *Value {
	return defaultManager.Create(amount, currency.TOP)
}

// FromTTD creates a new Value instance with TTD currency
func FromTTD(amount int64) *Value {
	return defaultManager.Create(amount, currency.TTD)
}

// FromTZS creates a new Value instance with TZS currency
func FromTZS(amount int64) *Value {
	return defaultManager.Create(amount, currency.TZS)
}

// FromUAH creates a new Value instance with UAH currency
func FromUAH(amount int64) *Value {
	return defaultManager.Create(amount, currency.UAH)
}

// FromUGX creates a new Value instance with UGX currency
func FromUGX(amount int64) *Value {
	return defaultManager.Create(amount, currency.UGX)
}

// FromUSN creates a new Value instance with USN currency
func FromUSN(amount int64) *Value {
	return defaultManager.Create(amount, currency.USN)
}

// FromUYI creates a new Value instance with UYI currency
func FromUYI(amount int64) *Value {
	return defaultManager.Create(amount, currency.UYI)
}

// FromUYU creates a new Value instance with UYU currency
func FromUYU(amount int64) *Value {
	return defaultManager.Create(amount, currency.UYU)
}

// FromUYW creates a new Value instance with UYW currency
func FromUYW(amount int64) *Value {
	return defaultManager.Create(amount, currency.UYW)
}

// FromUZS creates a new Value instance with UZS currency
func FromUZS(amount int64) *Value {
	return defaultManager.Create(amount, currency.UZS)
}

// FromVES creates a new Value instance with VES currency
func FromVES(amount int64) *Value {
	return defaultManager.Create(amount, currency.VES)
}

// FromVUV creates a new Value instance with VUV currency
func FromVUV(amount int64) *Value {
	return defaultManager.Create(amount, currency.VUV)
}

// FromWST creates a new Value instance with WST currency
func FromWST(amount int64) *Value {
	return defaultManager.Create(amount, currency.WST)
}

// FromXAF creates a new Value instance with XAF currency
func FromXAF(amount int64) *Value {
	return defaultManager.Create(amount, currency.XAF)
}

// FromXAG creates a new Value instance with XAG currency
func FromXAG(amount int64) *Value {
	return defaultManager.Create(amount, currency.XAG)
}

// FromXAU creates a new Value instance with XAU currency
func FromXAU(amount int64) *Value {
	return defaultManager.Create(amount, currency.XAU)
}

// FromXBA creates a new Value instance with XBA currency
func FromXBA(amount int64) *Value {
	return defaultManager.Create(amount, currency.XBA)
}

// FromXBB creates a new Value instance with XBB currency
func FromXBB(amount int64) *Value {
	return defaultManager.Create(amount, currency.XBB)
}

// FromXBC creates a new Value instance with XBC currency
func FromXBC(amount int64) *Value {
	return defaultManager.Create(amount, currency.XBC)
}

// FromXBD creates a new Value instance with XBD currency
func FromXBD(amount int64) *Value {
	return defaultManager.Create(amount, currency.XBD)
}

// FromXCD creates a new Value instance with XCD currency
func FromXCD(amount int64) *Value {
	return defaultManager.Create(amount, currency.XCD)
}

// FromXDR creates a new Value instance with XDR currency
func FromXDR(amount int64) *Value {
	return defaultManager.Create(amount, currency.XDR)
}

// FromXOF creates a new Value instance with XOF currency
func FromXOF(amount int64) *Value {
	return defaultManager.Create(amount, currency.XOF)
}

// FromXPD creates a new Value instance with XPD currency
func FromXPD(amount int64) *Value {
	return defaultManager.Create(amount, currency.XPD)
}

// FromXPF creates a new Value instance with XPF currency
func FromXPF(amount int64) *Value {
	return defaultManager.Create(amount, currency.XPF)
}

// FromXPT creates a new Value instance with XPT currency
func FromXPT(amount int64) *Value {
	return defaultManager.Create(amount, currency.XPT)
}

// FromXSU creates a new Value instance with XSU currency
func FromXSU(amount int64) *Value {
	return defaultManager.Create(amount, currency.XSU)
}

// FromXTS creates a new Value instance with XTS currency
func FromXTS(amount int64) *Value {
	return defaultManager.Create(amount, currency.XTS)
}

// FromXUA creates a new Value instance with XUA currency
func FromXUA(amount int64) *Value {
	return defaultManager.Create(amount, currency.XUA)
}

// FromXXX creates a new Value instance with XXX currency
func FromXXX(amount int64) *Value {
	return defaultManager.Create(amount, currency.XXX)
}

// FromYER creates a new Value instance with YER currency
func FromYER(amount int64) *Value {
	return defaultManager.Create(amount, currency.YER)
}

// FromZMW creates a new Value instance with ZMW currency
func FromZMW(amount int64) *Value {
	return defaultManager.Create(amount, currency.ZMW)
}

// FromZWL creates a new Value instance with ZWL currency
func FromZWL(amount int64) *Value {
	return defaultManager.Create(amount, currency.ZWL)
}
