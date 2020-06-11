package element

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/amitbet/dicom/dicomtag"
)

type JsonDicomVal struct {
	Name string `json:"name,omitempty"` //optional addition for better understanding
	Vr   string `json:"vr"`
	//Tag string `json:"Tag"`
	Value interface{} `json:"Value,omitempty"`
}
type JsonDicomPixelData struct {
	Frames []JsonDicomFrame `json:"frames"`
}
type JsonDicomFrame struct {
	FileOffset  int `json:"fileOffset"`
	SizeInBytes int `json:"sizeInBytes"`
}

func getElementValueAsJsonObj(el *Element, omitBinaryVals, addReadableNames, skipPrivateFields bool) (interface{}, error) {
	//var err error
	var val interface{}
	// ----- item in sequence:-----
	if el.Tag == dicomtag.Item {
		itemVal, err := getElementsAsJsonObj(el.Value, omitBinaryVals, addReadableNames, skipPrivateFields, nil)
		if err != nil {
			return nil, err
		}
		val = itemVal[0]
		//arrayOfDataSets = append(arrayOfDataSets, itemVal[0])
		// ----- pixel data:-----
	} else if el.Tag == dicomtag.PixelData {
		jPDdata := JsonDicomPixelData{}
		pdInfo := el.Value[0].(PixelDataInfo)
		for _, fr := range pdInfo.Frames {
			jFrame := JsonDicomFrame{
				FileOffset:  int(fr.FileOffset),
				SizeInBytes: fr.SizeInBytes,
			}
			jPDdata.Frames = append(jPDdata.Frames, jFrame)
		}
		val = jPDdata
		// ----- regular data fields:-----
	} else if el.VR == "SQ" {
		sqVal := []interface{}{}
		for _, sqEl := range el.Value {
			elObj, err := getElementValueAsJsonObj(sqEl.(*Element), omitBinaryVals, addReadableNames, skipPrivateFields)
			if err != nil {
				return nil, err
			}
			sqVal = append(sqVal, elObj)
		}
		val = sqVal
	} else if el.VR == "IS" {
		parsedValArr, err := parseIntStrArr(el.Value)
		val = parsedValArr
		if err != nil {
			val = el.Value
		}
	} else if el.VR == "DS" {
		parsedValArr, err := parseFloatStrArr(el.Value)
		val = parsedValArr
		if err != nil {
			val = el.Value
		}
	} else if (el.VR == "OB" || el.VR == "OW") && omitBinaryVals {
	} else {
		if el.Value == nil {
			return nil, nil
		}
		val = el.Value
	}
	return val, nil
}
func parseFloatStrArr(valArr []interface{}) ([]float64, error) {
	outArr := []float64{}
	for _, val := range valArr {
		fltVal, err := strconv.ParseFloat(val.(string), 64)
		if err != nil {
			return nil, err
		}

		outArr = append(outArr, fltVal)
	}
	return outArr, nil
}

func parseIntStrArr(valArr []interface{}) ([]int64, error) {
	outArr := []int64{}
	for _, val := range valArr {
		intVal, err := strconv.ParseInt(val.(string), 10, 64)
		if err != nil {
			return nil, err
		}

		outArr = append(outArr, intVal)
	}
	return outArr, nil
}

//getElementsAsJsonObj will return an object that represents the element array
func getElementsAsJsonObj(elements []interface{}, omitBinaryVals, addReadableNames, skipPrivateFields bool, tagsFilter map[string]interface{}) ([]interface{}, error) {

	jObjMap := map[string]interface{}{}
	var err error

	arrayOfDataSets := []interface{}{}
	for _, el1 := range elements {
		el, ok := el1.(*Element)
		if !ok {
			err = errors.New("Failed to cast value to element")
			return nil, err
		}

		//skip private tags if asked to do so
		if skipPrivateFields && el.Tag.Group%2 != 0 {
			continue
		}
		tagStr := fmt.Sprintf("%04X%04X", el.Tag.Group, el.Tag.Element)

		// skip tags not in filter in filter is not nil
		if tagsFilter != nil && len(tagsFilter) > 0 {
			_, wantedTag := tagsFilter[tagStr]

			// skip tag if not required specifically
			if !wantedTag {
				continue
			}
		}

		val, err := getElementValueAsJsonObj(el, omitBinaryVals, addReadableNames, skipPrivateFields)
		if err != nil {
			return nil, err
		}

		jObjVal := JsonDicomVal{
			Vr:    el.VR,
			Value: val,
		}

		if addReadableNames {
			name, _ := dicomtag.Find(el.Tag)
			jObjVal.Name = name.Name
		}
		jObjMap[tagStr] = jObjVal
	}
	arrayOfDataSets = append(arrayOfDataSets, jObjMap)
	return arrayOfDataSets, err
}

//GetDataSetAsJsonObj returns the DICOM dataset as an object for JSON serialization
func (ds *DataSet) GetDataSetAsJsonObj(omitBinaryVals, addReadableNames, skipPrivateFields bool) ([]interface{}, error) {
	elems := []interface{}{}
	for _, el := range ds.Elements {
		elems = append(elems, el)
	}
	return getElementsAsJsonObj(elems, omitBinaryVals, addReadableNames, skipPrivateFields, nil)
}

//GetDataSetAsJsonObjFiltered returns the DICOM dataset as an object for JSON serialization with tag filtering by given tags
func (ds *DataSet) GetDataSetAsJsonObjFiltered(omitBinaryVals, addReadableNames, skipPrivateFields bool, tags []dicomtag.Tag) ([]interface{}, error) {
	tagsFilter := map[string]interface{}{}
	for _, tag := range tags {
		tagStr := fmt.Sprintf("%04X%04X", tag.Group, tag.Element)
		tagsFilter[tagStr] = struct{}{}
	}

	elems := []interface{}{}
	for _, el := range ds.Elements {
		elems = append(elems, el)
	}
	return getElementsAsJsonObj(elems, omitBinaryVals, addReadableNames, skipPrivateFields, tagsFilter)
}

//GetDataSetAsJson marshals a dataset to a JSON string
func (ds *DataSet) GetDataSetAsJson(omitBinaryVals, addReadableNames, skipPrivateFields bool) (string, error) {
	jObj, err := ds.GetDataSetAsJsonObj(omitBinaryVals, addReadableNames, skipPrivateFields)
	if err != nil {
		return "", err
	}
	bytes, err := json.Marshal(jObj[0])
	return string(bytes), err
}

//GetDataSetAsJsonFiltered marshals a dataset to a JSON string with filtering by given tags
func (ds *DataSet) GetDataSetAsJsonFiltered(omitBinaryVals, addReadableNames, skipPrivateFields bool, tags []dicomtag.Tag) (string, error) {
	jObj, err := ds.GetDataSetAsJsonObjFiltered(omitBinaryVals, addReadableNames, skipPrivateFields, tags)
	if err != nil {
		return "", err
	}
	bytes, err := json.Marshal(jObj[0])
	return string(bytes), err
}

func GetDefaultMetadataTagFilter() []dicomtag.Tag {
	taglist := []dicomtag.Tag{}
	includedTags := []string{
		"00020002",
		"00020003",
		"00020010",
		"00020012",
		"00020013",
		"00020016",
		"00080005",
		"00080008",
		"00080012",
		"00080013",
		"00080016",
		"00080018",
		"00080020",
		"00080021",
		"00080022",
		"00080023",
		"00080030",
		"00080031",
		"00080032",
		"00080033",
		"00080050",
		"00080054",
		"00080060",
		"00080070",
		"00080080",
		"00080081",
		"00080090",
		"00081010",
		"00081020",
		"00081030",
		"00081040",
		"0008103E",
		"00081060",
		"00081070",
		"00081090",
		"00082111",
		"00090010",
		"00100010",
		"00100020",
		"00100030",
		"00100040",
		"00101001",
		"00101010",
		"001021B0",
		"00180010",
		"00180022",
		"00180050",
		"00180060",
		"00180088",
		"00180090",
		"00181020",
		"00181030",
		"00181040",
		"00181041",
		"00181046",
		"00181049",
		"00181100",
		"00181110",
		"00181111",
		"00181120",
		"00181130",
		"00181140",
		"00181150",
		"00181151",
		"00181160",
		"00181152",
		"00181170",
		"00181190",
		"00181210",
		"00185100",
		"00190010",
		"0020000D",
		"0020000E",
		"00200010",
		"00200011",
		"00200012",
		"00200013",
		"00200032",
		"00200037",
		"00200052",
		"00200060",
		"00201040",
		"00201041",
		"00210010",
		"00230010",
		"00270010",
		"00280002",
		"00280004",
		"00280010",
		"00280011",
		"00280030",
		"00280100",
		"00280101",
		"00280102",
		"00280103",
		"00280120",
		"00281050",
		"00281051",
		"00281052",
		"00281053",
		"00430010",
		"00450010",
		"00490010",
	}
	for _, str := range includedTags {
		tag := dicomtag.GetTagFromString(str)
		taglist = append(taglist, tag)
	}
	return taglist
}
