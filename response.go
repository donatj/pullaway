package pullaway

import "fmt"

type PushoverClientResponse struct {
	Status  int    `json:"status"`
	Request string `json:"request"`
	Errors  Errors `json:"errors,omitempty"`
}

type Errors struct {
	Name []string `json:"name"`
}

func (r *PushoverClientResponse) IsValid() bool {
	return r.Status == 1
}

func (r *PushoverClientResponse) Error() string {
	return fmt.Sprintf("status: %d, request: %s, errors: %#v", r.Status, r.Request, r.Errors)
}

type LoginResponse struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`

	PushoverClientResponse
}

type RegistrationResponse struct {
	ID string `json:"id"`

	PushoverClientResponse
}

type DownloadResponse struct {
	Messages []Messages `json:"messages"`
	User     User       `json:"user"`
	Device   Device     `json:"device"`

	PushoverClientResponse
}

func (r *DownloadResponse) MaxID() int64 {
	var max int64
	for _, m := range r.Messages {
		if m.ID > max {
			max = m.ID
		}
	}
	return max
}

type Messages struct {
	ID             int64  `json:"id"`
	IDStr          string `json:"id_str"`
	Message        string `json:"message"`
	App            string `json:"app"`
	Aid            int    `json:"aid"`
	AidStr         string `json:"aid_str"`
	Icon           string `json:"icon"`
	Date           int    `json:"date"`
	Priority       int    `json:"priority"`
	Acked          int    `json:"acked"`
	Umid           int64  `json:"umid"`
	UmidStr        string `json:"umid_str"`
	Title          string `json:"title"`
	DispatchedDate int    `json:"dispatched_date"`
	URL            string `json:"url,omitempty"`
	QueuedDate     int    `json:"queued_date,omitempty"`
}

type User struct {
	QuietHours        bool   `json:"quiet_hours"`
	IsAndroidLicensed bool   `json:"is_android_licensed"`
	IsIosLicensed     bool   `json:"is_ios_licensed"`
	IsDesktopLicensed bool   `json:"is_desktop_licensed"`
	Email             string `json:"email"`
	CreatedAt         int    `json:"created_at"`
	FirstEmailAlias   string `json:"first_email_alias"`
	ShowTipjar        string `json:"show_tipjar"`
	ShowTeamAd        string `json:"show_team_ad"`
}

type Device struct {
	Name                              string `json:"name"`
	EncryptionEnabled                 bool   `json:"encryption_enabled"`
	DefaultSound                      string `json:"default_sound"`
	AlwaysUseDefaultSound             bool   `json:"always_use_default_sound"`
	DefaultHighPrioritySound          string `json:"default_high_priority_sound"`
	AlwaysUseDefaultHighPrioritySound bool   `json:"always_use_default_high_priority_sound"`
	DismissalSyncEnabled              bool   `json:"dismissal_sync_enabled"`
}

type DeleteResponse struct {
	PushoverClientResponse
}
