package keeper_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	dg "github.com/babylonchain/babylon/testutil/datagen"
	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	bkeeper "github.com/babylonchain/babylon/x/btccheckpoint/keeper"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CheckpointInput struct {
	header                 string
	transactions           []string
	opReturnTransactionIdx int
	expectedOpReturnData   string
}

func InputsToSpvProofs(inputs []*CheckpointInput) []*btcctypes.BTCSpvProof {
	var spvs []*btcctypes.BTCSpvProof

	for _, input := range inputs {
		headerBytes, _ := hex.DecodeString(input.header)

		var txBytes [][]byte

		for _, t := range input.transactions {
			tbytes, _ := hex.DecodeString(t)
			txBytes = append(txBytes, tbytes)
		}

		spv, _ := dg.SpvProofFromHeaderAndTransactions(headerBytes, txBytes, uint(input.opReturnTransactionIdx))

		spvs = append(spvs, spv)
	}

	return spvs
}

func extractOpReturnFomInputs(inputs []*CheckpointInput) []byte {
	var opReturnData []byte

	for _, in := range inputs {
		dec, e := hex.DecodeString(in.expectedOpReturnData)

		if e != nil {
			panic("Inputs should contain valid hexencoded data")
		}

		opReturnData = append(opReturnData, dec...)
	}

	return opReturnData
}

var testData = []CheckpointInput{
	{
		"000000209e27bfdf755b434d57f3ce1a2045a6d1aba1832e07f209712700000000000000045e3c7c45090708ee3250519041bae1f9aa4df0cfe678f4fe63171a7c1d1edf9694b860fcff031a84f9c7b8",
		[]string{
			"020000000001010000000000000000000000000000000000000000000000000000000000000000ffffffff1503b2841e049694b8600134ea00030000ff00000000ffffffff02f4d597000000000017a914f5fb634163aee17801523fbfaee93d4baa6cf383870000000000000000266a24aa21a9ed95d6a57229d88063d35b29112f72bab9c0c03d7f4dd9e21b20f3c7f4cc9acb060120000000000000000000000000000000000000000000000000000000000000000000000000",
			"02000000000101e82231797b226eba2bce4ebec9e0052f76f7bca5dc36a51a06adacf17952daa10000000017160014bda2810a2886a5dea011c5757c74b49d7d4354defeffffff02ba511b00000000001976a91439f848db5fa3d62bed2960e8d171faab95ed299a88ac8de79d870000000017a914af4097ba2f04bfa21becae02d3581f9d00ca7a1187024730440220341fe850161413821f86327ab6a5ccfdec4af1b426569cddb9ac53489843456c02202b042e702545d0f2e6324e8f278417ee210ece39d334ddf54f8474af9b6fd7be01210293d3a3e5160d74990a17317c7b21aef661f245f69cbd6aeabc8f1432fdb7cdc9b1841e00",
			"02000000000101a6837fadd3f7a3395c8554e777343e48e046b34bd509615638277f4f10de5a780100000000feffffff02a49a05000000000017a9140b25fcf671a0e8f8be90e574d93d7eeb10bb4df9871cca45d60000000017a914eb579dca4d3be004509d6ccadcdee8e5dfd589b787024730440220325a54e10283495151baac99693acc02b0dfafc2012b8837c4945831003e593802201705713dd0c183ef1f80676a50e9e87fd157626e6ef1320794c101fb4e2a3750012102419b7dbbdba1486ce4f16dd65aa933462996d41d44904578c3ac0ecd5eb729b8b1841e00",
			"02000000000102d2fe7f7a75c381e2e7ec14f53f12ebe4d46a7762625555887a537fe0d4fcc72d0100000017160014322c1699d34b7d89c0bd1e7c7ce9687aa6eddb8201000000d3258e02da986d4bb4eb798582c32260d8433afeabf8c0648f1435263d63a12c0100000017160014322c1699d34b7d89c0bd1e7c7ce9687aa6eddb820100000001b28301000000000017a914334358eb21e5e497a50a2b797ad1a23a1a1fb45687024730440220458ba20f33b789393143116bfd41beef575ace5810bb53e1f7f9bcb736c55fcc02200ec6c7f951b3e77e7c3d9b4a6f058378d71c0cec44a162de5508dc90d019dc4f012102f17e4ed57d3570338a6f956d1e00bdd639dbd87f1b36a45a6a25ed75c01a988e024730440220417dc98167dffe9f8bce08a061268c7aca803c959207aaf1d04fc2f25092374402201cff35b71f5e3a2d8b2e9befc4ca410e7a65c79a93ee8ee99d10c7090255e85c012102f17e4ed57d3570338a6f956d1e00bdd639dbd87f1b36a45a6a25ed75c01a988e00000000",
			"020000000001011a78ddae1c56014a78e0d15d432d48e4f34922ef3bd9d6302ee7b9b6cdc0b6960000000000feffffff030000000000000000226a2000d2833d4071a235a7bf5aa87999e78587946aa89882ce8c118c97094b6e9defc011000000000000160014dd06c079ec009f9d758056c24981e92c60b4e04b3547000000000000160014971494f1a0eddce971754b93b65cf4e855cba03002473044022072c271dfd9077760ed390418cfd2bbc1de32684e81104334371462e561d9183f022027bf6885cc72d30d02debb0ac523a04cfcdc585f03e18a319cb0f0f9f2ed854c0121026925411a455fb52a628259a5c5833c82c5a6066a93a125668f8eff4eeace2d8f00000000",
			"02000000000101555fe667261847048d1a075eff879ba526ceccb3446c375a76a48dbd0bfd38710100000000feffffff020000000000000000226a20e81a786c066da715456ddf31d6d80a8689620a0bd722b18532ca481dc58ef60381040000000000001600142b4d3909e075fc496ba6fbc051fbe4c1671475670247304402205f2ea5c0092cde961afb500bb4446f7e6c1b134efa3774b8bc22f71699062f5d02207d5f85cca2e62ad86d32d1991dee4c0f75f9cae68f5ac227f49f26c63d4c15aa012102f15e9d9d6e35e805888aedd8b6469cf2607b784e39cb2bf90d7db2fc242a666700000000",
			"02000000000101b3354159353c9b86b9b3f5cf14699d1d92a12c4ff0430c154b908cb5073ca0430100000017160014a0642eba4abe84f01ac443727a92a83aacbd3d4cfeffffff02dbe807010000000017a9143f9055d003aedff712dffecdb90d8c590132261787a08601000000000017a914c6953606f7d751d8c1a956c888ce96ed97d7e09b870247304402203604f2167529c724bfe57f893652106ff71d81979b95c531028099c100c9746002202c12c7c0efed867556cb14bb310e7cc0b01c13037cb824feab00247482c1e7e1012102dcfd099718021f3792dc9222e65882722210982ba38a905b23236cbfd5e97bacb0841e00",
		},
		4,
		"00d2833d4071a235a7bf5aa87999e78587946aa89882ce8c118c97094b6e9def",
	},
	{
		"000000209e27bfdf755b434d57f3ce1a2045a6d1aba1832e07f209712700000000000000045e3c7c45090708ee3250519041bae1f9aa4df0cfe678f4fe63171a7c1d1edf9694b860fcff031a84f9c7b8",
		[]string{
			"020000000001010000000000000000000000000000000000000000000000000000000000000000ffffffff1503b2841e049694b8600134ea00030000ff00000000ffffffff02f4d597000000000017a914f5fb634163aee17801523fbfaee93d4baa6cf383870000000000000000266a24aa21a9ed95d6a57229d88063d35b29112f72bab9c0c03d7f4dd9e21b20f3c7f4cc9acb060120000000000000000000000000000000000000000000000000000000000000000000000000",
			"02000000000101e82231797b226eba2bce4ebec9e0052f76f7bca5dc36a51a06adacf17952daa10000000017160014bda2810a2886a5dea011c5757c74b49d7d4354defeffffff02ba511b00000000001976a91439f848db5fa3d62bed2960e8d171faab95ed299a88ac8de79d870000000017a914af4097ba2f04bfa21becae02d3581f9d00ca7a1187024730440220341fe850161413821f86327ab6a5ccfdec4af1b426569cddb9ac53489843456c02202b042e702545d0f2e6324e8f278417ee210ece39d334ddf54f8474af9b6fd7be01210293d3a3e5160d74990a17317c7b21aef661f245f69cbd6aeabc8f1432fdb7cdc9b1841e00",
			"02000000000101a6837fadd3f7a3395c8554e777343e48e046b34bd509615638277f4f10de5a780100000000feffffff02a49a05000000000017a9140b25fcf671a0e8f8be90e574d93d7eeb10bb4df9871cca45d60000000017a914eb579dca4d3be004509d6ccadcdee8e5dfd589b787024730440220325a54e10283495151baac99693acc02b0dfafc2012b8837c4945831003e593802201705713dd0c183ef1f80676a50e9e87fd157626e6ef1320794c101fb4e2a3750012102419b7dbbdba1486ce4f16dd65aa933462996d41d44904578c3ac0ecd5eb729b8b1841e00",
			"02000000000102d2fe7f7a75c381e2e7ec14f53f12ebe4d46a7762625555887a537fe0d4fcc72d0100000017160014322c1699d34b7d89c0bd1e7c7ce9687aa6eddb8201000000d3258e02da986d4bb4eb798582c32260d8433afeabf8c0648f1435263d63a12c0100000017160014322c1699d34b7d89c0bd1e7c7ce9687aa6eddb820100000001b28301000000000017a914334358eb21e5e497a50a2b797ad1a23a1a1fb45687024730440220458ba20f33b789393143116bfd41beef575ace5810bb53e1f7f9bcb736c55fcc02200ec6c7f951b3e77e7c3d9b4a6f058378d71c0cec44a162de5508dc90d019dc4f012102f17e4ed57d3570338a6f956d1e00bdd639dbd87f1b36a45a6a25ed75c01a988e024730440220417dc98167dffe9f8bce08a061268c7aca803c959207aaf1d04fc2f25092374402201cff35b71f5e3a2d8b2e9befc4ca410e7a65c79a93ee8ee99d10c7090255e85c012102f17e4ed57d3570338a6f956d1e00bdd639dbd87f1b36a45a6a25ed75c01a988e00000000",
			"020000000001011a78ddae1c56014a78e0d15d432d48e4f34922ef3bd9d6302ee7b9b6cdc0b6960000000000feffffff030000000000000000226a2000d2833d4071a235a7bf5aa87999e78587946aa89882ce8c118c97094b6e9defc011000000000000160014dd06c079ec009f9d758056c24981e92c60b4e04b3547000000000000160014971494f1a0eddce971754b93b65cf4e855cba03002473044022072c271dfd9077760ed390418cfd2bbc1de32684e81104334371462e561d9183f022027bf6885cc72d30d02debb0ac523a04cfcdc585f03e18a319cb0f0f9f2ed854c0121026925411a455fb52a628259a5c5833c82c5a6066a93a125668f8eff4eeace2d8f00000000",
			"02000000000101555fe667261847048d1a075eff879ba526ceccb3446c375a76a48dbd0bfd38710100000000feffffff020000000000000000226a20e81a786c066da715456ddf31d6d80a8689620a0bd722b18532ca481dc58ef60381040000000000001600142b4d3909e075fc496ba6fbc051fbe4c1671475670247304402205f2ea5c0092cde961afb500bb4446f7e6c1b134efa3774b8bc22f71699062f5d02207d5f85cca2e62ad86d32d1991dee4c0f75f9cae68f5ac227f49f26c63d4c15aa012102f15e9d9d6e35e805888aedd8b6469cf2607b784e39cb2bf90d7db2fc242a666700000000",
			"02000000000101b3354159353c9b86b9b3f5cf14699d1d92a12c4ff0430c154b908cb5073ca0430100000017160014a0642eba4abe84f01ac443727a92a83aacbd3d4cfeffffff02dbe807010000000017a9143f9055d003aedff712dffecdb90d8c590132261787a08601000000000017a914c6953606f7d751d8c1a956c888ce96ed97d7e09b870247304402203604f2167529c724bfe57f893652106ff71d81979b95c531028099c100c9746002202c12c7c0efed867556cb14bb310e7cc0b01c13037cb824feab00247482c1e7e1012102dcfd099718021f3792dc9222e65882722210982ba38a905b23236cbfd5e97bacb0841e00",
		},
		5,
		"e81a786c066da715456ddf31d6d80a8689620a0bd722b18532ca481dc58ef603",
	},
}

func TestSubmitValidNewCheckpoint(t *testing.T) {
	tests := []struct {
		inputs []*CheckpointInput
		epoch  uint64
	}{
		{
			[]*CheckpointInput{
				&testData[0],
				&testData[1],
			},
			1,
		},
	}

	for i, test := range tests {
		kDeep := 2
		wDeep := 4
		// here we will only have valid unconfirmed submissions
		lc := btcctypes.NewMockBTCLightClientKeeper(int64(kDeep) - 1)
		cc := btcctypes.NewMockCheckpointingKeeper(test.epoch)

		k, ctx := keepertest.BTCCheckpointKeeper(t, lc, cc, uint64(kDeep), uint64(wDeep))

		proofs := InputsToSpvProofs(test.inputs)

		pk, _ := dg.NewPV().GetPubKey()

		address := sdk.AccAddress(pk.Address().Bytes())

		msg := btcctypes.MsgInsertBTCSpvProof{
			Proofs:    proofs,
			Submitter: address.String(),
		}

		srv := bkeeper.NewMsgServerImpl(*k)

		sdkCtx := sdk.WrapSDKContext(ctx)

		_, err := srv.InsertBTCSpvProof(sdkCtx, &msg)

		if err != nil {
			t.Errorf("Unexpected message processing error: %v, testIdx: %d", err, i)
		}

		ed := k.GetEpochData(ctx, test.epoch)

		if len(ed.Key) == 0 {
			t.Errorf("There should be at least one key in epoch %d, testidx: %d", test.epoch, i)
		}

		if ed.Status != btcctypes.Submitted {
			t.Errorf("Epoch should be in submitted state after processing message")
		}

		expectedOpReturn := extractOpReturnFomInputs(test.inputs)

		if !bytes.Equal(expectedOpReturn, ed.RawCheckpoint) {
			t.Errorf("Epoch does not contain expected op return data")
		}

		submissionKey := ed.Key[0]

		submissionData := k.GetSubmissionData(ctx, *submissionKey)

		if submissionData == nil {
			t.Fatalf("Unexpected missing submission. Testidx: %d", i)
		}

		if submissionData.Epoch != test.epoch {
			t.Errorf("Submission data with invalid epoch. Testidx: %d", i)
		}

		allUnconfirmedSubmissions := k.GetAllUnconfirmedSubmissions(ctx)

		// TODO Add custom equal fo submission key and transaction key to check
		// it is expected key
		if len(allUnconfirmedSubmissions) == 0 {
			t.Errorf("Unexpected missing unconfirmed submissions. Idx: %d", i)
		}
	}
}

func TestStateTransitionOfValidSubmission(t *testing.T) {
	tests := []struct {
		inputs []*CheckpointInput
		epoch  uint64
		kDeep  uint64
		wDeep  uint64
	}{
		{
			[]*CheckpointInput{
				&testData[0],
				&testData[1],
			},
			1,
			2,
			4,
		},
	}

	for i, test := range tests {
		// here we will only have valid unconfirmed submissions
		lc := btcctypes.NewMockBTCLightClientKeeper(int64(test.kDeep) - 1)
		cc := btcctypes.NewMockCheckpointingKeeper(test.epoch)

		k, ctx := keepertest.BTCCheckpointKeeper(t, lc, cc, test.kDeep, test.wDeep)

		proofs := InputsToSpvProofs(test.inputs)

		pk, _ := dg.NewPV().GetPubKey()

		address := sdk.AccAddress(pk.Address().Bytes())

		msg := btcctypes.MsgInsertBTCSpvProof{
			Proofs:    proofs,
			Submitter: address.String(),
		}

		srv := bkeeper.NewMsgServerImpl(*k)

		sdkCtx := sdk.WrapSDKContext(ctx)

		_, err := srv.InsertBTCSpvProof(sdkCtx, &msg)

		if err != nil {
			t.Errorf("Unexpected message processing error: %v, testIdx: %d", err, i)
		}

		// TODO customs Equality for submission keys
		unc := k.GetAllUnconfirmedSubmissions(ctx)

		if len(unc) != 1 {
			t.Errorf("Unexpected missing unconfirmed submissions, testIdx: %d", i)
		}

		// Now we will return depth enough for moving submission to confirmed
		lc.SetDepth(int64(test.kDeep))

		// fire tip change callback
		k.OnTipChange(ctx)

		// TODO customs Equality for submission keys to check this are really keys
		// we are looking for
		unc = k.GetAllUnconfirmedSubmissions(ctx)
		conf := k.GetAllConfirmedSubmissions(ctx)

		if len(unc) != 0 {
			t.Errorf("Unexpected not promoted submission, testIdx: %d", i)
		}

		if len(conf) != 1 {
			t.Errorf("Unexpected missing confirmed submission, testIdx: %d", i)
		}

		ed := k.GetEpochData(ctx, test.epoch)

		if ed == nil || ed.Status != btcctypes.Confirmed {
			t.Errorf("Epoch Data missing of in unexpected state. TestIdx: %d", i)
		}

		lc.SetDepth(int64(test.wDeep))
		k.OnTipChange(ctx)

		unc = k.GetAllUnconfirmedSubmissions(ctx)
		conf = k.GetAllConfirmedSubmissions(ctx)
		fin := k.GetAllFinalizedSubmissions(ctx)

		if len(unc) != 0 {
			t.Errorf("Unexpected not promoted unconfirmed submission, testIdx: %d", i)
		}

		if len(conf) != 0 {
			t.Errorf("Unexpected not promoted confirmed submission, testIdx: %d", i)
		}

		if len(fin) != 1 {
			t.Errorf("Unexpected missing finalized submission, testIdx: %d", i)
		}

		ed = k.GetEpochData(ctx, test.epoch)

		if ed == nil || ed.Status != btcctypes.Finalized {
			t.Errorf("Epoch Data missing of in unexpected state. TestIdx: %d", i)
		}
	}

}
