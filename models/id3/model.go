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
	"strings"
	"time"

	"github.com/OpenMellon/go-xmp/models/dc"
	xmpbase "github.com/OpenMellon/go-xmp/models/xmp_base"
	xmpdm "github.com/OpenMellon/go-xmp/models/xmp_dm"
	xmprights "github.com/OpenMellon/go-xmp/models/xmp_rights"
	"github.com/OpenMellon/go-xmp/xmp"
)

var NsID3 = xmp.NewNamespace("id3", "http://id3.org/ns/2.4/", NewModel)

func init() {
	xmp.Register(NsID3, xmp.MusicMetadata)
}

func NewModel(name string) xmp.Model {
	if name == "id3" {
		return &ID3{}
	}
	return nil
}

func MakeModel(d *xmp.Document) (*ID3, error) {
	m, err := d.MakeModel(NsID3)
	if err != nil {
		return nil, err
	}
	x, _ := m.(*ID3)
	return x, nil
}

func FindModel(d *xmp.Document) *ID3 {
	if m := d.FindModel(NsID3); m != nil {
		return m.(*ID3)
	}
	return nil
}

func (m *ID3) Namespaces() xmp.NamespaceList {
	return xmp.NamespaceList{NsID3}
}

func (m *ID3) Can(nsName string) bool {
	return nsName == NsID3.GetName()
}

func (x *ID3) SyncModel(d *xmp.Document) error {
	return nil
}

func (x *ID3) SyncFromXMP(d *xmp.Document) error {
	if m := dc.FindModel(d); m != nil {
		if len(m.Title) > 0 {
			x.TitleDescription = m.Title.Default()
		}
		if len(m.Rights) > 0 {
			x.Copyright = m.Rights.Default()
		}
	}
	if base := xmpbase.FindModel(d); base != nil {
		if !base.CreateDate.IsZero() {
			x.RecordingTime = base.CreateDate
		}
	}
	if dm := xmpdm.FindModel(d); dm != nil {
		if dm.Artist != "" {
			x.LeadPerformer = dm.Artist
		}
		if dm.Album != "" {
			x.AlbumTitle = dm.Album
		}
		if dm.LogComment != "" {
			x.Comments = dm.LogComment
		}
		if dm.Genre != "" {
			if err := x.ContentType.UnmarshalText([]byte(dm.Genre)); err != nil {
				return err
			}
		}
		if dm.TrackNumber > 0 {
			x.TrackNumber = TrackNum{
				Track: dm.TrackNumber,
			}
		}
		if dm.PartOfCompilation.Value() {
			x.IsCompilation = True
		} else {
			x.IsCompilation = False
		}
		if dm.Composer != "" {
			x.Composer = dm.Composer
		}
		if dm.Engineer != "" {
			x.ModifiedBy = dm.Engineer
		}
		if dm.DiscNumber > 0 {
			x.PartOfASet = TrackNum{
				Track: dm.DiscNumber,
			}
		}
		if dm.Lyrics != "" {
			x.UnsynchronizedLyrics = strings.Split(dm.Lyrics, "\n")
		}
	}
	if rights := xmprights.FindModel(d); rights != nil {
		if rights.WebStatement != "" {
			x.CopyrightInformation = xmp.Url(rights.WebStatement)
		}
	}

	return nil
}

func (x ID3) SyncToXMP(d *xmp.Document) error {
	// pick default language for multi-lang XMP properties
	lang := x.Language

	// DC gets title (TIT2) and copyright (TCOP)
	if x.TitleDescription != "" || x.Copyright != "" {
		m, err := dc.MakeModel(d)
		if err != nil {
			return err
		}
		if x.TitleDescription != "" && len(m.Title) == 0 {
			m.Title.AddDefault(lang, x.TitleDescription)
		}
		if x.Copyright != "" && len(m.Rights) == 0 {
			m.Rights.AddDefault(lang, x.Copyright)
		}
	}

	// XMP base gets creation date
	if x.RecordingTime.IsZero() && !x.Date_v23.IsZero() && !x.Time_v23.IsZero() {
		_, mm, dd := x.Date_v23.Value().Date()
		h, m, s := x.Time_v23.Value().Clock()
		x.RecordingTime = xmp.Date(time.Date(x.Year_v23, mm, dd, h, m, s, 0, nil))
	}
	if !x.RecordingTime.IsZero() {
		base, err := xmpbase.MakeModel(d)
		if err != nil {
			return err
		}
		base.CreateDate = x.RecordingTime
	}

	// XmpDM gets always created
	dm, err := xmpdm.MakeModel(d)
	if err != nil {
		return err
	}
	if x.Comments != "" {
		dm.LogComment = x.Comments
	}
	if x.AlbumTitle != "" {
		dm.Album = x.AlbumTitle
	}
	dm.PartOfCompilation = xmp.Bool(x.IsCompilation.Value())
	if x.Composer != "" {
		dm.Composer = x.Composer
	}
	if len(x.ContentType) > 0 {
		dm.Genre = x.ContentType.String()
	}
	if x.LeadPerformer != "" {
		dm.Artist = x.LeadPerformer
	}
	if x.ModifiedBy != "" {
		dm.Engineer = x.ModifiedBy
	}
	if !x.PartOfASet.IsZero() {
		dm.DiscNumber = x.PartOfASet.Track
	}
	if !x.TrackNumber.IsZero() {
		dm.TrackNumber = x.TrackNumber.Track
	}
	if len(x.UnsynchronizedLyrics) > 0 {
		dm.Lyrics = strings.Join(x.UnsynchronizedLyrics, "\n")
	}
	// XmpRights gets copyright
	if x.CopyrightInformation != "" {
		rights, err := xmprights.MakeModel(d)
		if err != nil {
			return err
		}
		rights.WebStatement = string(x.CopyrightInformation)
	}
	return nil
}
