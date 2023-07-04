package model

import "encoding/json"

func UnmarshalLogin(data []byte) (Login, error) {
	var r Login
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Login) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Login struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func UnmarshalAccessToken(data []byte) (AccessToken, error) {
	var r AccessToken
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AccessToken) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type AccessToken struct {
	TokenType             string `json:"token_type"`
	AccessToken           string `json:"access_token"`
	RefreshToken          string `json:"refresh_token"`
	ExpiresIn             int64  `json:"expires_in"`
	RefreshTokenExpiresIn int64  `json:"refresh_token_expires_in"`
}

func UnmarshalVideo(data []byte) (Video, error) {
	var r Video
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Video) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Video struct {
	Video VideoClass `json:"video"`
}

type VideoClass struct {
	ID        int64  `json:"id"`
	UUID      string `json:"uuid"`
	ShortUUID string `json:"shortUUID"`
}

type Media struct {
	Title       string
	Description string
	FilePath    string
}

func UnmarshalMetadata(data []byte) (Metadata, error) {
	var r Metadata
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Metadata) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Metadata struct {
	Streams []Stream `json:"streams"`
	Format  Format   `json:"format"`
}

type Format struct {
	Filename       string     `json:"filename"`
	NbStreams      int64      `json:"nb_streams"`
	NbPrograms     int64      `json:"nb_programs"`
	FormatName     string     `json:"format_name"`
	FormatLongName string     `json:"format_long_name"`
	StartTime      string     `json:"start_time"`
	Duration       string     `json:"duration"`
	Size           string     `json:"size"`
	BitRate        string     `json:"bit_rate"`
	ProbeScore     int64      `json:"probe_score"`
	Tags           FormatTags `json:"tags"`
}

type FormatTags struct {
	MajorBrand       string `json:"major_brand"`
	MinorVersion     string `json:"minor_version"`
	CompatibleBrands string `json:"compatible_brands"`
	CreationTime     string `json:"creation_time"`
}

type Stream struct {
	Index            int64            `json:"index"`
	CodecName        string           `json:"codec_name"`
	CodecLongName    string           `json:"codec_long_name"`
	Profile          string           `json:"profile"`
	CodecType        string           `json:"codec_type"`
	CodecTagString   string           `json:"codec_tag_string"`
	CodecTag         string           `json:"codec_tag"`
	Width            int64            `json:"width"`
	Height           int64            `json:"height"`
	CodedWidth       int64            `json:"coded_width"`
	CodedHeight      int64            `json:"coded_height"`
	ClosedCaptions   int64            `json:"closed_captions"`
	FilmGrain        int64            `json:"film_grain"`
	HasBFrames       int64            `json:"has_b_frames"`
	PixFmt           string           `json:"pix_fmt"`
	Level            int64            `json:"level"`
	ColorRange       string           `json:"color_range"`
	ColorSpace       string           `json:"color_space"`
	ColorTransfer    string           `json:"color_transfer"`
	ColorPrimaries   string           `json:"color_primaries"`
	ChromaLocation   string           `json:"chroma_location"`
	FieldOrder       string           `json:"field_order"`
	Refs             int64            `json:"refs"`
	IsAVC            string           `json:"is_avc"`
	NalLengthSize    string           `json:"nal_length_size"`
	ID               string           `json:"id"`
	RFrameRate       string           `json:"r_frame_rate"`
	AvgFrameRate     string           `json:"avg_frame_rate"`
	TimeBase         string           `json:"time_base"`
	StartPts         int64            `json:"start_pts"`
	StartTime        string           `json:"start_time"`
	DurationTs       int64            `json:"duration_ts"`
	Duration         string           `json:"duration"`
	BitRate          string           `json:"bit_rate"`
	BitsPerRawSample string           `json:"bits_per_raw_sample"`
	NbFrames         string           `json:"nb_frames"`
	ExtradataSize    int64            `json:"extradata_size"`
	Disposition      map[string]int64 `json:"disposition"`
	Tags             StreamTags       `json:"tags"`
}

type StreamTags struct {
	CreationTime string `json:"creation_time"`
	Language     string `json:"language"`
	HandlerName  string `json:"handler_name"`
	VendorID     string `json:"vendor_id"`
	Encoder      string `json:"encoder"`
}
