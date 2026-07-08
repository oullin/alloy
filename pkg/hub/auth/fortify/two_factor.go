package fortify

import (
	"net/http"
	"time"

	"github.com/oullin/alloy/pkg/hub/auth/twofactor"
	cauth "github.com/oullin/alloy/pkg/hub/contracts/auth"
)

// TwoFactorConfig controls two-factor endpoint output.
type TwoFactorConfig struct {
	Issuer      string
	AccountName func(user cauth.User) string
	Window      int
}

// NewEnableTwoFactorHandler generates a secret and plaintext recovery codes.
func NewEnableTwoFactorHandler(guard cauth.Guard, persist TwoFactorUpdater, cfg TwoFactorConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, twoFactor, ok := twoFactorUser(w, r, guard)

		if !ok {
			return
		}

		secret, err := twofactor.GenerateSecret()

		if err != nil {
			writeError(w, http.StatusInternalServerError, "two-factor setup failed")

			return
		}

		recoveryCodes, err := twofactor.GenerateRecoveryCodes(8)

		if err != nil {
			writeError(w, http.StatusInternalServerError, "two-factor setup failed")

			return
		}

		twoFactor.SetTwoFactorSecret(secret)
		twoFactor.SetTwoFactorRecoveryCodes(twofactor.HashRecoveryCodes(recoveryCodes))
		twoFactor.SetTwoFactorEnabled(false)
		twoFactor.SetTwoFactorConfirmedAt(nil)

		if err := persistTwoFactor(r, twoFactor, persist); err != nil {
			writeError(w, http.StatusUnprocessableEntity, "two-factor setup failed")

			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"secret":         secret,
			"otpauth_url":    twofactor.OTPAuthURL(twoFactorIssuer(cfg), twoFactorAccountName(cfg, user), secret),
			"recovery_codes": recoveryCodes,
		})
	}
}

// NewConfirmTwoFactorHandler verifies a TOTP code and enables 2FA.
func NewConfirmTwoFactorHandler(guard cauth.Guard, persist TwoFactorUpdater, cfg TwoFactorConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, twoFactor, ok := twoFactorUser(w, r, guard)

		if !ok {
			return
		}

		input, err := readInput(r)

		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid request")

			return
		}

		code := stringInput(input, "code")

		if code == "" {
			writeValidation(w, "code", "required")

			return
		}

		if !twofactor.Verify(twoFactor.GetTwoFactorSecret(), code, time.Now(), twoFactorWindow(cfg)) {
			writeValidation(w, "code", "invalid code")

			return
		}

		now := time.Now()
		twoFactor.SetTwoFactorEnabled(true)
		twoFactor.SetTwoFactorConfirmedAt(&now)

		if err := persistTwoFactor(r, twoFactor, persist); err != nil {
			writeError(w, http.StatusUnprocessableEntity, "two-factor confirmation failed")

			return
		}

		writeOK(w, "two-factor authentication enabled")
	}
}

// NewDisableTwoFactorHandler disables 2FA and clears secrets.
func NewDisableTwoFactorHandler(guard cauth.Guard, persist TwoFactorUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, twoFactor, ok := twoFactorUser(w, r, guard)

		if !ok {
			return
		}

		twoFactor.SetTwoFactorEnabled(false)
		twoFactor.SetTwoFactorSecret("")
		twoFactor.SetTwoFactorRecoveryCodes(nil)
		twoFactor.SetTwoFactorConfirmedAt(nil)

		if err := persistTwoFactor(r, twoFactor, persist); err != nil {
			writeError(w, http.StatusUnprocessableEntity, "two-factor disable failed")

			return
		}

		writeNoContent(w)
	}
}

// NewRegenerateRecoveryCodesHandler returns new plaintext recovery codes once.
func NewRegenerateRecoveryCodesHandler(guard cauth.Guard, persist TwoFactorUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, twoFactor, ok := twoFactorUser(w, r, guard)

		if !ok {
			return
		}

		recoveryCodes, err := twofactor.GenerateRecoveryCodes(8)

		if err != nil {
			writeError(w, http.StatusInternalServerError, "recovery code generation failed")

			return
		}

		twoFactor.SetTwoFactorRecoveryCodes(twofactor.HashRecoveryCodes(recoveryCodes))

		if err := persistTwoFactor(r, twoFactor, persist); err != nil {
			writeError(w, http.StatusUnprocessableEntity, "recovery code generation failed")

			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"recovery_codes": recoveryCodes})
	}
}

func twoFactorUser(w http.ResponseWriter, r *http.Request, guard cauth.Guard) (cauth.User, cauth.TwoFactorUser, bool) {
	user, err := guard.User(r.Context())

	if err != nil || user == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated")

		return nil, nil, false
	}

	twoFactor, ok := user.(cauth.TwoFactorUser)

	if !ok {
		writeError(w, http.StatusUnprocessableEntity, "two-factor authentication is not supported")

		return nil, nil, false
	}

	return user, twoFactor, true
}

func persistTwoFactor(r *http.Request, user cauth.TwoFactorUser, persist TwoFactorUpdater) error {
	if persist == nil {
		return nil
	}

	return persist(r.Context(), user)
}

func twoFactorIssuer(cfg TwoFactorConfig) string {
	if cfg.Issuer == "" {
		return "Alloy"
	}

	return cfg.Issuer
}

func twoFactorAccountName(cfg TwoFactorConfig, user cauth.User) string {
	if cfg.AccountName != nil {
		return cfg.AccountName(user)
	}

	return user.GetAuthIdentifier()
}

func twoFactorWindow(cfg TwoFactorConfig) int {
	if cfg.Window < 0 {
		return 0
	}

	if cfg.Window == 0 {
		return 1
	}

	return cfg.Window
}
