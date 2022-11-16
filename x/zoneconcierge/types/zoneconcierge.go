package types

func (ci *ChainInfo) TryToUpdateHeader(header *IndexedHeader) {
	if ci.LatestHeader != nil {
		// ensure the header is the latest one
		// NOTE: submitting an old header is considered acceptable in IBC-Go (see Case_valid_past_update),
		// but the chain info indexer will not record such old header since it's not the latest one
		if ci.LatestHeader.Height > header.Height {
			return
		}
	}
	ci.LatestHeader = header
}

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
