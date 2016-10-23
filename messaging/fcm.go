package messaging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Endpoint is the URL where all Firebase Cloud Messages are sent to.
const Endpoint = "https://fcm.googleapis.com/fcm/send"

// Messaing represents the Firebase Cloud Messaging service.
type Messaing struct {
	apiKey string
	client *http.Client
}

// New returns a new instance of the Firebase Cloud Message service.
func New(apiKey string, client *http.Client) *Messaing {
	if client == nil {
		client = http.DefaultClient
	}

	return &Messaing{
		apiKey: apiKey,
		client: client,
	}
}

// Send posts a message to the Firebase Cloud Messaging service.
func (f *Messaing) Send(msg Message) (*Response, error) {
	b, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", Endpoint, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("key=%s", f.apiKey))
	req.Header.Add("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response Response
	return &response, json.NewDecoder(resp.Body).Decode(&response)
}

// Response represents the FCM server's response to the application
// server's sent message.
type Response struct {
	MulticastID  int64    `json:"multicast_id"`
	Success      int      `json:"success"`
	Failure      int      `json:"failure"`
	CanonicalIDs int      `json:"canonical_ids"`
	Results      []Result `json:"results"`
}

// Failures collects and returns all failed results.
func (r *Response) Failures() []Result {
	results := make([]Result, r.Failure)
	resultIndex := 0

	for _, result := range r.Results {
		if result.Error == "" {
			continue
		}
		results[resultIndex] = result
		resultIndex++
	}

	return results
}

// Result represents the status of a processed message.
type Result struct {
	MessageID      string `json:"message_id"`
	RegistrationID string `json:"registration_id"`
	Error          string `json:"error"`
}

// Message represents the payload that is sent to the Firebase Cloud Messaging
// service.
// Ref https://firebase.google.com/docs/cloud-messaging/http-server-ref#downstream-http-messages-json
// Message represents list of targets, options, and payload for HTTP JSON messages.
type Message struct {
	// Token specifies the recipient of a message. The value must be a
	// registration token, notification key, or topic. Do not set this
	// field when sending to multiple topics. See condition.
	Token string `json:"to,omitempty"`

	// RegistrationIDs specifies a list of devices (registration tokens, or IDs)
	// receiving a multicast message. It must contain at least 1 and at most 1000 registration tokens.
	//
	// Use this parameter only for multicast messaging, not for single recipients.
	RegistrationIDs []string `json:"registration_ids,omitempty"`

	// Condition specifies a logical expression of conditions that determine the message target.
	// Supported condition: Topic, formatted as "'yourTopic' in topics". This value is case-insensitive.
	// Supported operators: &&, ||. Maximum two operators per topic message supported.
	Condition string `json:"condition,omitempty"`

	// CollapseKey identifies a group of messages (e.g., with collapse_key: "Updates Available")
	// that can be collapsed, so that only the last message gets sent when delivery can be resumed.
	// This is intended to avoid sending too many of the same messages when the device comes back online or becomes active.
	//
	// Note that there is no guarantee of the order in which messages get sent.
	//
	// Note: A maximum of 4 different collapse keys is allowed at any given time.
	// This means a FCM connection server can simultaneously store 4 different send-to-sync
	// messages per client app. If you exceed this number, there is no guarantee which 4 collapse
	// keys the FCM connection server will keep.

	CollapseKey string `json:"collapse_key,omitempty"`

	// Priority sets the priority of the message. Valid values are "normal" and "high."
	// On iOS, these correspond to APNs priorities 5 and 10.
	// By default, messages are sent with normal priority. Normal priority optimizes
	// the client app's battery consumption and should be used unless immediate delivery
	// is required. For messages with normal priority, the app may receive the message with unspecified delay.
	//
	// When a message is sent with high priority, it is sent immediately, and the app can
	// wake a sleeping device and open a network connection to your server.
	Priority string `json:"priority,omitempty"`

	// ContentAvailable signifies whether or not content is avaliable as part of the message.
	//
	// On iOS, use this field to represent content-available in the APNs payload.
	// When a notification or message is sent and this is set to true, an inactive client
	// app is awoken.
	//
	// On Android, data messages wake the app by default.
	//
	// On Chrome, currently not supported.
	ContentAvailable bool `json:"content_available,omitempty"`

	// TimeToLive specifies how long (in seconds) the message should be kept in
	// FCM storage if the device is offline. The maximum time to live supported
	// is 4 weeks, and the default value is 4 weeks
	TimeToLive int `json:"time_to_live,omitempty"`

	// RestrictedPackageName specifies the package name of the application where the
	// registration tokens must match in order to receive the message.
	RestrictedPackageName string `json:"restricted_package_name,omitempty"`

	// DryRun allows developers to test a request without actuallin sending it.
	//
	// The default value is false.
	DryRun bool `json:"dry_run,omitempty"`

	// Notification specifies the predefined, user-visible key-value pairs of
	// the notification payload. See Notification payload support for detail.
	Notification *Notification `json:"notification,omitempty"`

	// Data specifies the custom key-value pairs of the message's payload.
	//
	// For example, with data:{"score":"3x1"}:
	//
	// On iOS, if the message is sent via APNS, it represents the custom data fields.
	// If it is sent via FCM connection server, it would be represented as key value
	// dictionary in AppDelegate application:didReceiveRemoteNotification:.
	//
	// On Android, this would result in an intent extra named score with the string value 3x1.
	//
	// The key should not be a reserved word ("from" or any word starting with "google" or "gcm").
	// Do not use any of the words defined in this table (such as collapse_key).
	//
	// Values in string types are recommended. You have to convert values in objects or other
	// non-string data types (e.g., integers or booleans) to string.
	Data *Data `json:"data,omitempty"`
}

// Notification specifies the predefined, user-visible key-value pairs
// of the notification payload
type Notification struct {
	// Title indicates the notification title. This field is not visible on iOS phones and tablets.
	Title string `json:"title,omitempty"`

	// Body indicates the notification body text.
	Body string `json:"body,omitempty"`

	// Icon indicates notification icon.
	// On Android sets value to myicon for drawable resource myicon.
	// On Web it is the URL for a notification icon.
	Icon string `json:"icon,omitempty"`

	// Sound indicates the sound to play when the device receives a notification.
	// Sound files can be in the main bundle of the client app or in the Library/Sounds
	// folder of the app's data container.
	// Only used for iOS and Android
	Sound string `json:"sound,omitempty"`

	// Badge indicates the badge on the client app home icon.
	// Only used for iOS
	Badge string `json:"badge,omitempty"`

	// Tag indicates whether each notification results in a new entry in the notification drawer on Android.
	// If not set, each request creates a new notification.
	// If set, and a notification with the same tag is already being shown, the new
	// notification replaces the existing one in the notification drawer.
	// Only used for Android
	Tag string `json:"tag,omitempty"`

	// Color indicates color of the icon, expressed in #rrggbb format
	// Only used for Android
	Color string `json:"color,omitempty"`

	// ClickAction indicates the action associated with a user click on the notification.
	// On iOS it corresponds to category in the APNs payload.
	// On Android when this is set, an activity with a matching intent filter is launched when user clicks the notification.
	// On Web all URL values, secure HTTPS is required.
	ClickAction string `json:"click_action,omitempty"`

	// BodyLocKey indicates the key to the body string for localization.
	// On iOS it corresponds to "loc-key" in the APNs payload.
	// On Android it is the key in the app's string resources when populating this value.
	BodyLocKey string `json:"body_loc_key,omitempty"`

	// BodyLocArgs indicates the string value to replace format specifiers in the body string for localization.
	// On iOS it corresponds to "loc-args" in the APNs payload.
	// On Android for more information, see http://developer.android.com/guide/topics/resources/string-resource.html#FormattingAndStyling
	BodyLocArgs string `json:"body_loc_args,omitempty"`

	// TitleLocKey indicates the key to the title string for localization.
	// On iOS it corresponds to "title-loc-key" in the APNs payload.
	// On Android it is the key in the app's string resources when populating this value.
	TitleLocKey string `json:"title_loc_key,omitempty"`

	// TitleLocArgs indicates the string value to replace format specifiers in the title string for localization.
	// On iOS it corresponds to "title-loc-args" in the APNs payload.
	// On Android for more information, see http://developer.android.com/guide/topics/resources/string-resource.html#FormattingAndStyling
	TitleLocArgs string `json:"title_loc_args,omitempty"`
}

// Data specifies the custom key-value pairs of the message's payload.
type Data map[string]interface{}
