// Copyright 2012 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/acolwell/mse-tools/webm"
)

type InitSegment struct {
	Offset int64
	Size   int64
}

type MediaSegment struct {
	Offset   int64
	Size     int64
	Timecode float64
	Frames   []*webm.BlockInfo
}

type JSONManifest struct {
	Type                 string
	Duration             float64
	StartDate            time.Time
	Init                 *InitSegment
	Media                []*MediaSegment
	TimecodeScale        uint64
	DefaultTrackDuration uint64
}

func (jm *JSONManifest) ToJSON() string {
	str := "{\n"
	str += "  \"type\": \"" + strings.Replace(jm.Type, "\"", "\\\"", -1) + "\",\n"
	if jm.Duration == -1 {
		str += "  \"live\": true, \n"
	} else {
		str += fmt.Sprintf("  \"duration\": %f,\n", jm.Duration)
	}

	str += fmt.Sprintf("  \"timescale\": { \"1\": %d },\n", jm.TimecodeScale)

	if !jm.StartDate.IsZero() {
		str += "  \"startDate\": " + jm.StartDate.Format(time.RFC3339Nano) + ", \n"
	}

	str += fmt.Sprintf("  \"init\": { \"offset\": %d, \"size\": %d},\n",
		jm.Init.Offset,
		jm.Init.Size)
	str += "  \"media\": [\n"
	for i := range jm.Media {
		m := jm.Media[i]
		str += fmt.Sprintf("    { \"offset\": %d, \"size\": %d, \"type\": \"TODO\", \"timecode\": %f",
			m.Offset,
			m.Size,
			m.Timecode)

		if strings.HasPrefix(jm.Type, "video") {
			// Add idr_frames to JSON output:
			idrFrames := []int{}
			for i, frame := range m.Frames {
				if isKeyFrame(frame) {
					idrFrames = append(idrFrames, i)
				}
			}
			buf, _ := json.Marshal(idrFrames)
			str += fmt.Sprintf(", \"idr_frames\": {\"1\": %s }", buf)
		}

		str += fmt.Sprintf(", \"tick_count\": {\"1\": %d}", len(m.Frames))
		str += fmt.Sprintf(", \"tick_size\": {\"1\": %f}", float64(jm.DefaultTrackDuration)/1000)

		str += fmt.Sprint(" }")

		if i+1 != len(jm.Media) {
			str += ","
		}
		str += "\n"
	}
	str += "  ]\n"
	str += "}\n"
	return str
}

func isKeyFrame(b *webm.BlockInfo) bool {
	return (b.Flags & 0x80) != 0
}

func NewJSONManifest() *JSONManifest {
	return &JSONManifest{Type: "",
		Duration: -1,
		Init:     nil,
		Media:    []*MediaSegment{},
	}
}
