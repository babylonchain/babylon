package configurer

import "time"

type setupFn func(configurer Configurer) error

func baseSetup(configurer Configurer) error {
	if err := configurer.RunValidators(); err != nil {
		return err
	}
	return nil
}

func withIBC(setupHandler setupFn) setupFn {
	return func(configurer Configurer) error {
		if err := setupHandler(configurer); err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
		if err := configurer.RunHermesRelayerIBC(); err != nil {
			return err
		}

		return nil
	}
}

func withPhase2IBC(setupHandler setupFn) setupFn {
	return func(configurer Configurer) error {
		if err := setupHandler(configurer); err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
		// Instantiate contract on (CZ-like) chain B
		if err := configurer.InstantiateBabylonContract(); err != nil {
			return err
		}

		if err := configurer.RunHermesRelayerIBC(); err != nil {
			return err
		}

		return nil
	}
}

func withPhase2RlyIBC(setupHandler setupFn) setupFn {
	return func(configurer Configurer) error {
		if err := setupHandler(configurer); err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
		// Instantiate contract on (CZ-like) chain B
		if err := configurer.InstantiateBabylonContract(); err != nil {
			return err
		}

		if err := configurer.RunCosmosRelayerIBC(); err != nil {
			return err
		}

		return nil
	}
}

func withIBCTransferChannel(setupHandler setupFn) setupFn {
	return func(configurer Configurer) error {
		if err := setupHandler(configurer); err != nil {
			return err
		}
		time.Sleep(5 * time.Second)

		if err := configurer.RunIBCTransferChannel(); err != nil {
			return err
		}

		return nil
	}
}
