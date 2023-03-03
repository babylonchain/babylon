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
		time.Sleep(5 * time.Second)
		if err := configurer.RunIBC(); err != nil {
			return err
		}

		return nil
	}
}
