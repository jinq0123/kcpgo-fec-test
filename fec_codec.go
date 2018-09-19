package main

type FecCodec struct {
	dec *fecDecoder
	enc *fecEncoder
}

func NewFecCodec() *FecCodec {
	// FEC keeps rxFECMulti* (dataShard+parityShard) ordered packets in memory
	const rxFECMulti = 3

	// FEC codec initialization
	return &FecCodec{
		dec: newFECDecoder(rxFECMulti*(dataShards+parityShards), dataShards, parityShards),
		enc: newFECEncoder(dataShards, parityShards, 0),
	}
}
