#!/usr/bin/env bash

# QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv
go build -o highway && ./highway -privatekey CAMSeTB3AgEBBCDtIHJcnRKCWVtitn0gkRTHlKvJCvSO12XVtzHna3oSEqAKBggqhkjOPQMBB6FEA0IABKQXV3mHcxNSmL3n4mtWTO4vNP2IuPvizYngBGxf6Fx9cCJhYUYH8r+Plp40dVcz53iXFxbtxIU3Z5oIVVOsYvI= -support_shards all --loglevel info

# QmRfp2xEWdRwagEfxVXTJee6xqaJnWTPogpKKdutq3FHNQ
# go build -o highway && ./highway -privatekey CAMSeTB3AgEBBCBPHYHjWDGEG4irMbsODtYwv4+PSrrpEA0mOCdymU5sKKAKBggqhkjOPQMBB6FEA0IABJxbC13BH4WFm3ILpGCn6xxjn8Fsqg6GOmufDqJ+o1kRN1cCeYT0Z3agP50iD3TOx55FNCinWhqEUDW3at2+vAU= -support_shards all -proxy_port 9331 -bootstrap dummy
