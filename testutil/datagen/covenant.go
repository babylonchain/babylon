package datagen

import (
	"github.com/babylonchain/babylon/btcstaking"
	asig "github.com/babylonchain/babylon/crypto/schnorr-adaptor-signature"
	bbn "github.com/babylonchain/babylon/types"
	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/wire"
)

func GenCovenantAdaptorSigs(
	covenantSKs []*btcec.PrivateKey,
	valPKs []*btcec.PublicKey,
	fundingTx *wire.MsgTx,
	pkScriptPath []byte,
	slashingTx *bstypes.BTCSlashingTx,
) ([]*bstypes.CovenantAdaptorSignatures, error) {
	covenantSigs := []*bstypes.CovenantAdaptorSignatures{}
	for _, covenantSK := range covenantSKs {
		covMemberSigs := &bstypes.CovenantAdaptorSignatures{
			CovPk:       bbn.NewBIP340PubKeyFromBTCPK(covenantSK.PubKey()),
			AdaptorSigs: [][]byte{},
		}
		for _, valPK := range valPKs {
			encKey, err := asig.NewEncryptionKeyFromBTCPK(valPK)
			if err != nil {
				return nil, err
			}
			covenantSig, err := slashingTx.EncSign(fundingTx, 0, pkScriptPath, covenantSK, encKey)
			if err != nil {
				return nil, err
			}
			covMemberSigs.AdaptorSigs = append(covMemberSigs.AdaptorSigs, covenantSig.MustMarshal())
		}
		covenantSigs = append(covenantSigs, covMemberSigs)
	}

	return covenantSigs, nil
}

func GenCovenantUnbondingSigs(covenantSKs []*btcec.PrivateKey, stakingTx *wire.MsgTx, stakingOutIdx uint32, unbondingPkScriptPath []byte, unbondingTx *wire.MsgTx) ([]*schnorr.Signature, error) {
	sigs := []*schnorr.Signature{}
	for i := range covenantSKs {
		sig, err := btcstaking.SignTxWithOneScriptSpendInputStrict(
			unbondingTx,
			stakingTx,
			stakingOutIdx,
			unbondingPkScriptPath,
			covenantSKs[i],
		)
		if err != nil {
			return nil, err
		}
		sigs = append(sigs, sig)
	}
	return sigs, nil
}
