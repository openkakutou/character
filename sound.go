package character

// Sound is a single decoded MUGEN/Ikemen GO sound effect belonging to a
// Character's SoundFile, keyed the same way a .cns PlaySnd controller
// addresses it: by (Group, Sample). Decoding itself is delegated to the
// external github.com/openkakutou/snd module — see loadSounds/loadSoundsFromBytes.
type Sound struct {
	// Group is the sound group index this sound belongs to.
	Group int `json:"group"`
	// Sample is this sound's sample index within its Group.
	Sample int `json:"sample"`
	// SampleRate is the decoded audio's sample rate, in Hz.
	SampleRate int `json:"sampleRate"`
	// Channels is the number of interleaved audio channels.
	Channels int `json:"channels"`
	// BitsPerSample is the original source bit depth (8 or 16), before
	// normalization to 16-bit PCM.
	BitsPerSample int `json:"bitsPerSample"`
	// PCM is the decoded audio, interleaved by channel and normalized to
	// signed 16-bit samples regardless of the original source bit depth.
	PCM []int16 `json:"pcm"`
}

// SoundGroup is a collection of Sounds that share the same group index —
// mirrors sff.SpriteGroup's shape, applied to the sound domain instead of
// sprites.
type SoundGroup struct {
	// Index is the sound group index shared by every Sound in Sounds.
	Index int `json:"index"`
	// Sounds is the ordered collection of sounds belonging to this group.
	Sounds []Sound `json:"sounds"`
}
