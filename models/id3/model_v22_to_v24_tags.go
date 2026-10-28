// Copyright (c) 2017-2018 Alexander Eichhorn
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// ID3 v2.2, v2.3 & v2.4
// http://id3.org/

// TODO:
// - almost all TextUnmarshalers are unimplemented due to missing samples
// - embedded XMP may be carried inside a PRIV frame with XMP owner identifier
// - multi-language text fields (requires samples)
// - multi-lang comments using CommentArray type (requires samples)

// FIXME: Version 2.4 of the specification prescribes that all text fields
// (the fields that start with a T, except for TXXX) can contain multiple
// values separated by a null character. The null character varies by
// character encoding.

// Package id3 implements the ID3v2.4 metadata standard for audio files.
package id3

import (
	"fmt"

	"github.com/OpenMellon/go-xmp/xmp"
)

var V22_TO_V24_TAGS map[string]string = map[string]string{
	"BUF": "RBUF", //   Recommended buffer size
	"CNT": "PCNT", //   Play counter
	"COM": "COMM", //   Comments
	"CRA": "AENC", //   Audio encryption
	"CRM": "ENCR", //   Encrypted meta frame
	"ETC": "ETCO", //   Event timing codes
	"EQU": "EQU2", //   Equalization
	"GEO": "GEOB", //   General encapsulated object
	"IPL": "TIPL", //   Involved people list
	"LNK": "LINK", //   Linked information
	"MCI": "MCDI", //   Music CD Identifier
	"MLL": "MLLT", //   MPEG location lookup table
	"PIC": "APIC", //   Attached picture
	"POP": "POPM", //   Popularimeter
	"REV": "RVRB", //   Reverb
	"RVA": "RVA2", //   Relative volume adjustment
	"SLT": "SYLT", //   Synchronized lyric/text
	"STC": "SYTC", //   Synced tempo codes
	"TAL": "TALB", //   Album/Movie/Show title
	"TBP": "TBPM", //   BPM (Beats Per Minute)
	"TCM": "TCOM", //   Composer
	"TCO": "TMED", //   Content type
	"TCR": "TCOP", //   Copyright message
	"TDA": "TDRC", // # Date, arguably this could also be release date TDRL
	"TDY": "TDLY", //   Playlist delay
	"TEN": "TENC", //   Encoded by
	"TFT": "TFLT", //   File type
	"TIM": "TDRC", // # Time, arguably this could also be release date TDRL
	"TKE": "TKEY", //   Initial key
	"TLA": "TLAN", //   Language(s)
	"TLE": "TLEN", //   Length
	"TMT": "TMED", //   Media type
	"TOA": "TOPE", //   Original artist(s)/performer(s)
	"TOF": "TOFN", //   Original filename
	"TOL": "TOLY", //   Original Lyricist(s)/text writer(s)
	"TOR": "TORY", //   Original release year (v2.3-)
	"TOT": "TOAL", //   Original album/Movie/Show title
	"TP1": "TPE1", //   Lead artist(s)/Lead performer(s)/Soloist(s)/Performing group
	"TP2": "TPE2", //   Band/Orchestra/Accompaniment
	"TP3": "TPE3", //   Conductor/Performer refinement
	"TP4": "TPE4", //   Interpreted, remixed, or otherwise modified by
	"TPA": "TPOS", //   Part of a set
	"TPB": "TPUB", //   Publisher
	"TRC": "TSRC", //   ISRC (International Standard Recording Code)
	"TRD": "TDRC", //   Recording dates
	"TRK": "TRCK", //   Track number/Position in set
	"TSI": "TSIZ", //   Size (v2.3-)
	"TSS": "TSSE", //   Software/hardware and settings used for encoding
	"TT1": "TIT1", //   Content group description
	"TT2": "TIT2", //   Title/Songname/Content description
	"TT3": "TIT3", //   Subtitle/Description refinement
	"TXT": "TEXT", //   Lyricist/text writer
	"TXX": "TXXX", //   User defined text information frame
	"TYE": "TYER", //   Year (v2.3-)
	"UFI": "UFID", //   Unique file identifier
	"ULT": "USLT", //   Unsychronized lyric/text transcription
	"WAF": "WOAF", //   Official audio file webpage
	"WAR": "WOAR", //   Official artist/performer webpage
	"WAS": "WOAS", //   Official audio source webpage
	"WCM": "WCOM", //   Commercial information
	"WCP": "WCOP", //   Copyright/Legal information
	"WPB": "WPUB", //   Publishers official webpage
	"WXX": "WXXX", //   User defined URL link frame
	"TCP": "TCMP", //   iTunes compilation
	"TST": "TSOT", //   iTunes title sort order
	"TS2": "TSO2", //   iTunes album_artist_sort
	"TSA": "TSOA", //   iTunes album_sort
	"TSP": "TSOP", //   iTunes artist_sort
	"TSC": "TSOC", //   iTunes composer_sort
}

// tag can be a four letter v2.3/v2.4 code or a 3 letter v2.2 code
func mapTagToV24(tag string) (string, error) {
	if len(tag) > 4 {
		if t, ok := V11_TO_V24_TAGS[tag]; !ok {
			return tag, fmt.Errorf("id3: unknown v1 tag '%s'", tag)
		} else {
			tag = t
		}
	} else if len(tag) < 4 {
		if t, ok := V22_TO_V24_TAGS[tag]; !ok {
			return tag, fmt.Errorf("id3: unknown v2.2 tag '%s'", tag)
		} else {
			tag = t
		}
	}
	return tag, nil
}

func (x *ID3) CanTag(tag string) bool {
	t, err := mapTagToV24(tag)
	if err != nil {
		return false
	}
	_, err = xmp.GetNativeField(x, t)
	return err == nil
}

func (x *ID3) GetTag(tag string) (string, error) {
	t, err := mapTagToV24(tag)
	if err != nil {
		return "", err
	}
	if v, err := xmp.GetNativeField(x, t); err != nil {
		return "", fmt.Errorf("%s: %v", NsID3.GetName(), err)
	} else {
		return v, nil
	}
}

func (x *ID3) SetTag(tag, value string) error {
	t, err := mapTagToV24(tag)
	if err != nil {
		return err
	}
	if err := xmp.SetNativeField(x, t, value); err != nil {
		return fmt.Errorf("%s: %v", NsID3.GetName(), err)
	}
	return nil
}

// Returns tag value as string for the selected language. Falls back to
// default language when lang is empty.
func (x *ID3) GetLocaleTag(lang string, tag string) (string, error) {
	t, err := mapTagToV24(tag)
	if err != nil {
		return "", err
	}
	if val, err := xmp.GetLocaleField(x, lang, t); err != nil {
		return "", fmt.Errorf("%s: %v", NsID3.GetName(), err)
	} else {
		return val, nil
	}
}

func (x *ID3) SetLocaleTag(lang string, tag, value string) error {
	if err := xmp.SetLocaleField(x, lang, tag, value); err != nil {
		return fmt.Errorf("%s: %v", NsID3.GetName(), err)
	}
	return nil
}

// Lists all non-empty tags.
func (x *ID3) ListTags() (xmp.TagList, error) {
	if l, err := xmp.ListNativeFields(x); err != nil {
		return nil, fmt.Errorf("%s: %v", NsID3.GetName(), err)
	} else {
		return l, nil
	}
}
