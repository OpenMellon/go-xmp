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

package id3

const (
	ChannelTypeOther       ChannelType = 0x00 // Other
	ChannelTypeMaster      ChannelType = 0x01 // Master volume
	ChannelTypeFrontRight  ChannelType = 0x02 // Front right
	ChannelTypeFrontLeft   ChannelType = 0x03 // Front left
	ChannelTypeBackRight   ChannelType = 0x04 // Back right
	ChannelTypeBackLeft    ChannelType = 0x05 // Back left
	ChannelTypeFrontCentre ChannelType = 0x06 // Front centre
	ChannelTypeBackCentre  ChannelType = 0x07 // Back centre
	ChannelTypeSubwoofer   ChannelType = 0x08 // Subwoofer
)

type EqualizationMethod byte

const (
	// No interpolation is made. A jump from one adjustment level to
	// another occurs in the middle between two adjustment points.
	EqualizationMethodBand EqualizationMethod = 0x00
	// Interpolation between adjustment points is linear.
	EqualizationMethodLinear EqualizationMethod = 0x01
)

type PictureType byte

const (
	PictureTypeOther         PictureType = 0x00 // Other
	PictureTypeFileIcon32    PictureType = 0x01 // 32x32 pixels 'file icon' (PNG only)
	PictureTypeFileIcon      PictureType = 0x02 // Other file icon
	PictureTypeFrontCover    PictureType = 0x03 // Cover (front)
	PictureTypeBackCover     PictureType = 0x04 // Cover (back)
	PictureTypeLeaflet       PictureType = 0x05 // Leaflet page
	PictureTypeMedia         PictureType = 0x06 // Media (e.g. label side of CD)
	PictureTypePerformer     PictureType = 0x07 // Lead artist/lead performer/soloist
	PictureTypeArtist        PictureType = 0x08 // Artist/performer
	PictureTypeConductor     PictureType = 0x09 // Conductor
	PictureTypeBand          PictureType = 0x0A // Band/Orchestra
	PictureTypeComposer      PictureType = 0x0B // Composer
	PictureTypeWriter        PictureType = 0x0C // Lyricist/text writer
	PictureTypeLocation      PictureType = 0x0D // Recording Location
	PictureTypeRecording     PictureType = 0x0E // During recording
	PictureTypePerformance   PictureType = 0x0F // During performance
	PictureTypeSnapshot      PictureType = 0x10 // Movie/video screen capture
	PictureTypeFinish        PictureType = 0x11 // A bright coloured fish
	PictureTypeIllustration  PictureType = 0x12 // Illustration
	PictureTypeArtistLogo    PictureType = 0x13 // Band/artist logotype
	PictureTypePublisherLogo PictureType = 0x14 // Publisher/Studio logotype
)

type DeliveryMethod byte

const (
	DeliveryMethodOther        DeliveryMethod = 0x00 // Other
	DeliveryMethodStdCD        DeliveryMethod = 0x01 // Standard CD album with other songs
	DeliveryMethodCompressedCD DeliveryMethod = 0x02 // Compressed audio on CD
	DeliveryMethodDownload     DeliveryMethod = 0x03 // File over the Internet
	DeliveryMethodStream       DeliveryMethod = 0x04 // Stream over the Internet
	DeliveryMethodNoteSheets   DeliveryMethod = 0x05 // As note sheets
	DeliveryMethodBook         DeliveryMethod = 0x06 // As note sheets in a book with other sheets
	DeliveryMethodOtherMedia   DeliveryMethod = 0x07 // Music on other media
	DeliveryMethodMerchandise  DeliveryMethod = 0x08 // Non-musical merchandise
)
