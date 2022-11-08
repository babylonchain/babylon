package types

func (ci *ChainInfo) TryToUpdateForkHeader(header *IndexedHeader) {
	if len(ci.LatestForks.Headers) == 0 { // no fork at the moment
		ci.LatestForks.Headers = append(ci.LatestForks.Headers, header)
	} else if ci.LatestForks.Headers[0].Height == header.Height { // there exists fork headers at the same height
		ci.LatestForks.Headers = append(ci.LatestForks.Headers, header)
	} else if ci.LatestForks.Headers[0].Height < header.Height { // this fork header is newer than the previous one
		ci.LatestForks = &Forks{
			Headers: []*IndexedHeader{header},
		}
	}
}
