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

// XMP EXIF Mapping for Exif 2.3 metadata CIPA DC-010-2012
//
// Exif Spec
// Exif 2.3 http://www.cipa.jp/std/documents/e/DC-010-2012_E.pdf
// Exif 2.3 http://www.cipa.jp/std/documents/e/DC-008-2012_E.pdf
// Exif 2.3.1 http://www.cipa.jp/std/documents/e/DC-008-Translation-2016-E.pdf
//
// see https://www.media.mit.edu/pia/Research/deepview/exif.html
// for a very good explanation of tags and ifd's

// Package exif implements the Exif 2.3.1 metadata standard as defined in CIPA DC-008-2016.
package exif

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/OpenMellon/go-xmp/models/dc"
	"github.com/OpenMellon/go-xmp/models/ps"
	"github.com/OpenMellon/go-xmp/models/tiff"
	xmpbase "github.com/OpenMellon/go-xmp/models/xmp_base"
	"github.com/OpenMellon/go-xmp/xmp"
)

func (x *ExifInfo) SyncFromXMP(d *xmp.Document) error {
	if !x.DateTimeDigitizedXMP.IsZero() {
		x.DateTimeDigitized = Date(x.DateTimeDigitizedXMP.Value())
		x.SubSecTimeDigitized = strconv.Itoa(x.DateTimeDigitizedXMP.Value().Nanosecond())
	}

	if len(x.ISOSpeedRatings) > 0 {
		x.ExPhotographicSensitivity = x.ISOSpeedRatings[0]
	}

	// TODO
	// convert GPS from XMP to exif values

	if m := dc.FindModel(d); m != nil {
		x.ArtistXMP = m.Creator
		x.Artist = strings.Join(m.Creator, ",")
		x.ImageDescriptionXMP = m.Description
		x.ImageDescription = m.Description.Default()
		x.CopyrightXMP = m.Rights
		x.Copyright = m.Rights.Default()
	}
	if base := xmpbase.FindModel(d); base != nil {
		if !base.ModifyDate.IsZero() {
			x.DateTimeXMP = base.ModifyDate
			x.DateTime = Date(base.ModifyDate.Value())
			x.SubSecTime = strconv.Itoa(base.ModifyDate.Value().Nanosecond())
		}
		if base.CreatorTool.IsZero() {
			x.SoftwareXMP = base.CreatorTool
			x.Software = base.CreatorTool.String()
		}
	}
	if tff := tiff.FindModel(d); tff != nil {
		x.BitsPerSample = tff.BitsPerSample
		x.Compression = tff.Compression
		x.ImageLength = tff.ImageLength
		x.ImageWidth = tff.ImageWidth
		x.Make = tff.Make
		x.Model = tff.Model
		x.Orientation = tff.Orientation
		x.PhotometricInterpretation = tff.PhotometricInterpretation
		x.PlanarConfiguration = tff.PlanarConfiguration
		x.PrimaryChromaticities = tff.PrimaryChromaticities
		x.ReferenceBlackWhite = tff.ReferenceBlackWhite
		x.ResolutionUnit = tff.ResolutionUnit
		x.SamplesPerPixel = tff.SamplesPerPixel
		x.TransferFunction = tff.TransferFunction
		x.WhitePoint = tff.WhitePoint
		x.XResolution = tff.XResolution
		x.YCbCrCoefficients = tff.YCbCrCoefficients
		x.YCbCrPositioning = tff.YCbCrPositioning
		x.YCbCrSubSampling = tff.YCbCrSubSampling
		x.YResolution = tff.YResolution
	}
	if m := ps.FindModel(d); m != nil {
		if !m.DateCreated.IsZero() {
			x.DateTimeOriginalXMP = m.DateCreated
			x.DateTimeOriginal = Date(m.DateCreated.Value())
			x.SubSecTimeOriginal = strconv.Itoa(m.DateCreated.Value().Nanosecond())
		}
	}
	return nil
}

func (x *ExifInfo) SyncToXMP(d *xmp.Document) error {
	// convert dates, text and GPS properties, ignore errors
	if !x.DateTimeOriginal.IsZero() {
		x.DateTimeOriginalXMP, _ = convertDateToXMP(x.DateTimeOriginal, x.SubSecTimeOriginal)
	}
	if !x.DateTimeDigitized.IsZero() {
		x.DateTimeDigitizedXMP, _ = convertDateToXMP(x.DateTimeDigitized, x.SubSecTimeDigitized)
	}
	if !x.DateTime.IsZero() {
		x.DateTimeXMP, _ = convertDateToXMP(x.DateTime, x.SubSecTime)
	}
	if x.Artist != "" {
		x.ArtistXMP = xmp.StringList(strings.Split(x.Artist, ";"))
	}
	if x.ImageDescription != "" && len(x.ImageDescriptionXMP) == 0 {
		x.ImageDescriptionXMP = xmp.NewAltString(x.ImageDescription)
	}
	if x.Copyright != "" && len(x.CopyrightXMP) == 0 {
		x.CopyrightXMP = xmp.NewAltString(x.Copyright)
	}
	if x.Software != "" && x.SoftwareXMP.IsZero() {
		x.SoftwareXMP = xmp.AgentName(x.Software)
	}

	if len(x.ISOSpeedRatings) > 0 {
		x.ExPhotographicSensitivity = x.ISOSpeedRatings[0]
	}

	// convert GPS coordinates
	if v, err := convertGPStoXMP(x.GPSLatitude, x.GPSLatitudeRef); err == nil && v != "" {
		x.GPSLatitudeCoord = v
	}
	if v, err := convertGPStoXMP(x.GPSLongitude, x.GPSLongitudeRef); err == nil && v != "" {
		x.GPSLongitudeCoord = v
	}
	if v, err := convertGPStoXMP(x.GPSDestLatitude, x.GPSDestLatitudeRef); err == nil && v != "" {
		x.GPSDestLatitudeCoord = v
	}
	if v, err := convertGPStoXMP(x.GPSDestLongitude, x.GPSDestLongitudeRef); err == nil && v != "" {
		x.GPSDestLongitudeCoord = v
	}
	if !x.GPSDateStamp.IsZero() {
		x.GPSTimeStampXMP = convertGPSTimestamp(x.GPSDateStamp, x.GPSTimeStamp)
	}
	return nil
}

func (x *ExifInfo) CanTag(tag string) bool {
	_, err := xmp.GetNativeField(x, tag)
	return err == nil
}

func (x *ExifInfo) GetTag(tag string) (string, error) {
	tag = strings.ToLower(tag)
	if v, err := xmp.GetNativeField(x, tag); err != nil {
		return "", fmt.Errorf("exif: %v", err)
	} else {
		return v, nil
	}
}

func (x *ExifInfo) SetTag(tag, value string) error {
	tag = strings.ToLower(tag)
	if err := xmp.SetNativeField(x, tag, value); err != nil {
		return fmt.Errorf("exif: %v", err)
	}
	return nil
}

func (x *ExifInfo) GetLocaleTag(lang string, tag string) (string, error) {
	tag = strings.ToLower(tag)
	if val, err := xmp.GetLocaleField(x, lang, tag); err != nil {
		return "", fmt.Errorf("exif: %v", err)
	} else {
		return val, nil
	}
}

func (x *ExifInfo) SetLocaleTag(lang string, tag, value string) error {
	tag = strings.ToLower(tag)
	if err := xmp.SetLocaleField(x, lang, tag, value); err != nil {
		return fmt.Errorf("exif: %v", err)
	}
	return nil
}

type ExifEXInfo struct {
	InteroperabilityIndex     InteropMode       `exif:"0x0001" xmp:"exifEX:InteroperabilityIndex"`
	InteroperabilityVersion   string            `exif:"0x0002" xmp:"exifEX:InteroperabilityVersion"`
	RelatedImageFileFormat    string            `exif:"0x1000" xmp:"exifEX:RelatedImageFileFormat"`
	RelatedImageWidth         int               `exif:"0x1001" xmp:"exifEX:RelatedImageWidth"`
	RelatedImageLength        int               `exif:"0x1002" xmp:"exifEX:RelatedImageLength"`
	PhotographicSensitivity   int               `exif:"0x8827" xmp:"exifEX:PhotographicSensitivity"`
	SensitivityType           SensitivityType   `exif:"0x8830" xmp:"exifEX:SensitivityType"`
	StandardOutputSensitivity int               `exif:"0x8831" xmp:"exifEX:StandardOutputSensitivity"`
	RecommendedExposureIndex  int               `exif:"0x8832" xmp:"exifEX:RecommendedExposureIndex"`
	ISOSpeed                  int               `exif:"0x8833" xmp:"exifEX:ISOSpeed"`
	ISOSpeedLatitudeyyy       int               `exif:"0x8834" xmp:"exifEX:ISOSpeedLatitudeyyy"`
	ISOSpeedLatitudezzz       int               `exif:"0x8835" xmp:"exifEX:ISOSpeedLatitudezzz"`
	CameraOwnerName           string            `exif:"0xa430" xmp:"exifEX:CameraOwnerName"`
	BodySerialNumber          string            `exif:"0xa431" xmp:"exifEX:BodySerialNumber"`
	LensSpecification         xmp.RationalArray `exif:"0xa432" xmp:"exifEX:LensSpecification"` // 4 components: min/max [mm] + min/max [f-stop]
	LensMake                  string            `exif:"0xa433" xmp:"exifEX:LensMake"`
	LensModel                 string            `exif:"0xa434" xmp:"exifEX:LensModel"`
	LensSerialNumber          string            `exif:"0xa435" xmp:"exifEX:LensSerialNumber"`
	Gamma                     xmp.Rational      `exif:"0xa500" xmp:"exifEX:Gamma"`
}

func (m *ExifEXInfo) Namespaces() xmp.NamespaceList {
	return xmp.NamespaceList{NsExifEX}
}

func (m *ExifEXInfo) Can(nsName string) bool {
	return nsName == NsExifEX.GetName()
}

func (x *ExifEXInfo) SyncModel(d *xmp.Document) error {
	return nil
}

func (x *ExifEXInfo) SyncFromXMP(d *xmp.Document) error {
	return nil
}

func (x ExifEXInfo) SyncToXMP(d *xmp.Document) error {
	return nil
}

func (x *ExifEXInfo) CanTag(tag string) bool {
	_, err := xmp.GetNativeField(x, tag)
	return err == nil
}

func (x *ExifEXInfo) GetTag(tag string) (string, error) {
	if v, err := xmp.GetNativeField(x, tag); err != nil {
		return "", fmt.Errorf("%s: %v", NsExifEX.GetName(), err)
	} else {
		return v, nil
	}
}

func (x *ExifEXInfo) SetTag(tag, value string) error {
	if err := xmp.SetNativeField(x, tag, value); err != nil {
		return fmt.Errorf("%s: %v", NsExifEX.GetName(), err)
	}
	return nil
}

// XMP-EXIF Extra Mappings (XMP only)
// The schema namespace URI is http://ns.adobe.com/exif/1.0/aux/
// The preferred schema namespace prefix is aux
//
// Adobe-defined auxiliary EXIF tags.  This namespace existed in the XMP
// specification until it was dropped in 2012, presumably due to the
// introduction of the EXIF 2.3 for XMP specification and the exifEX namespace
// at this time.
type ExifAuxInfo struct {
	ApproximateFocusDistance                           xmp.Rational `xmp:"aux:ApproximateFocusDistance"`                           // 315/10
	DistortionCorrectionAlreadyApplied                 xmp.Bool     `xmp:"aux:DistortionCorrectionAlreadyApplied"`                 //
	SerialNumber                                       string       `xmp:"aux:SerialNumber"`                                       // 2481231346
	LensInfo                                           string       `xmp:"aux:LensInfo"`                                           // 10/1 22/1 0/0 0/0
	LensDistortInfo                                    string       `xmp:"aux:LensDistortInfo"`                                    //
	Lens                                               string       `xmp:"aux:Lens"`                                               // EF-S10-22mm f/3.5-4.5 USM
	LensID                                             string       `xmp:"aux:LensID"`                                             // 235
	LensSerialNumber                                   string       `xmp:"aux:LensSerialNumber"`                                   //
	ImageNumber                                        int          `xmp:"aux:ImageNumber"`                                        // 0
	IsMergedHDR                                        xmp.Bool     `xmp:"aux:IsMergedHDR"`                                        //
	IsMergedPanorama                                   xmp.Bool     `xmp:"aux:IsMergedPanorama"`                                   //
	LateralChromaticAberrationCorrectionAlreadyApplied xmp.Bool     `xmp:"aux:LateralChromaticAberrationCorrectionAlreadyApplied"` //
	VignetteCorrectionAlreadyApplied                   xmp.Bool     `xmp:"aux:VignetteCorrectionAlreadyApplied"`                   //
	FlashCompensation                                  xmp.Rational `xmp:"aux:FlashCompensation"`                                  // 0/1
	Firmware                                           string       `xmp:"aux:Firmware"`                                           // 1.2.5
	OwnerName                                          string       `xmp:"aux:OwnerName"`                                          // unknown
}

func (m *ExifAuxInfo) Namespaces() xmp.NamespaceList {
	return xmp.NamespaceList{NsExifAux}
}

func (m *ExifAuxInfo) Can(nsName string) bool {
	return nsName == NsExifAux.GetName()
}

func (x *ExifAuxInfo) SyncModel(d *xmp.Document) error {
	return nil
}

func (x *ExifAuxInfo) SyncFromXMP(d *xmp.Document) error {
	return nil
}

func (x ExifAuxInfo) SyncToXMP(d *xmp.Document) error {
	return nil
}

func (x *ExifAuxInfo) CanTag(tag string) bool {
	_, err := xmp.GetNativeField(x, tag)
	return err == nil
}

func (x *ExifAuxInfo) GetTag(tag string) (string, error) {
	if v, err := xmp.GetNativeField(x, tag); err != nil {
		return "", fmt.Errorf("%s: %v", NsExifAux.GetName(), err)
	} else {
		return v, nil
	}
}

func (x *ExifAuxInfo) SetTag(tag, value string) error {
	if err := xmp.SetNativeField(x, tag, value); err != nil {
		return fmt.Errorf("%s: %v", NsExifAux.GetName(), err)
	}
	return nil
}
