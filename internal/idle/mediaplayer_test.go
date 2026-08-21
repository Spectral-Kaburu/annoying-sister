package idle

import "testing"

func TestPrettyPlayerName(t *testing.T) {
	cases := map[string]string{
		"org.mpris.MediaPlayer2.vlc":                "Vlc",
		"org.mpris.MediaPlayer2.vlc.instance1234":    "Vlc",
		"org.mpris.MediaPlayer2.mpv":                 "Mpv",
		"org.mpris.MediaPlayer2.chromium.instance42": "Chromium",
		"org.mpris.MediaPlayer2.":                    "something",
	}

	for input, want := range cases {
		if got := prettyPlayerName(input); got != want {
			t.Errorf("prettyPlayerName(%q) = %q, want %q", input, got, want)
		}
	}
}
