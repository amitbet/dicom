package element

import (
	"encoding/json"
	"errors"
	"fmt"

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

func getElementValueAsJsonObj(el *Element, omitBinaryVals, addReadableNames bool) (interface{}, error) {
	//var err error
	var val interface{}
	// ----- item in sequence:-----
	if el.Tag == dicomtag.Item {
		itemVal, err := getElementsAsJsonObj(el.Value, omitBinaryVals, addReadableNames, nil)
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
			elObj, err := getElementValueAsJsonObj(sqEl.(*Element), omitBinaryVals, addReadableNames)
			if err != nil {
				return nil, err
			}
			sqVal = append(sqVal, elObj)
		}
		val = sqVal
	} else if (el.VR == "OB" || el.VR == "OW") && omitBinaryVals {
		val = ""
	} else if len(el.Value) == 1 {
		val = el.Value
		// ----- sequences :-----
	}
	return val, nil
}

//getElementsAsJsonObj will return an object that represents the element array
func getElementsAsJsonObj(elements []interface{}, omitBinaryVals, addReadableNames bool, tagsFilter map[string]interface{}) ([]interface{}, error) {

	jObjMap := map[string]interface{}{}
	var err error

	arrayOfDataSets := []interface{}{}
	for _, el1 := range elements {
		el, ok := el1.(*Element)
		if !ok {
			err = errors.New("Failed to cast value to element")
			return nil, err
		}
		tagStr := fmt.Sprintf("%04X%04X", el.Tag.Group, el.Tag.Element)

		if tagsFilter != nil && len(tagsFilter) > 0 {
			_, wantedTag := tagsFilter[tagStr]

			// skip tag if not required specifically
			if !wantedTag {
				continue
			}
		}

		val, err := getElementValueAsJsonObj(el, omitBinaryVals, addReadableNames)
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
func (ds *DataSet) GetDataSetAsJsonObj(omitBinaryVals, addReadableNames bool) ([]interface{}, error) {
	elems := []interface{}{}
	for _, el := range ds.Elements {
		elems = append(elems, el)
	}
	return getElementsAsJsonObj(elems, omitBinaryVals, addReadableNames, nil)
}

//GetDataSetAsJsonObjFiltered returns the DICOM dataset as an object for JSON serialization with tag filtering by given tags
func (ds *DataSet) GetDataSetAsJsonObjFiltered(omitBinaryVals, addReadableNames bool, tags []dicomtag.Tag) ([]interface{}, error) {
	tagsFilter := map[string]interface{}{}
	for _, tag := range tags {
		tagStr := fmt.Sprintf("%04X%04X", tag.Group, tag.Element)
		tagsFilter[tagStr] = struct{}{}
	}

	elems := []interface{}{}
	for _, el := range ds.Elements {
		elems = append(elems, el)
	}
	return getElementsAsJsonObj(elems, omitBinaryVals, addReadableNames, tagsFilter)
}

//GetDataSetAsJson marshals a dataset to a JSON string
func (ds *DataSet) GetDataSetAsJson(omitBinaryVals, addReadableNames bool) (string, error) {
	jObj, err := ds.GetDataSetAsJsonObj(omitBinaryVals, addReadableNames)
	if err != nil {
		return "undefined", err
	}
	bytes, err := json.Marshal(jObj[0])
	return string(bytes), err
}

//GetDataSetAsJsonFiltered marshals a dataset to a JSON string with filtering by given tags
func (ds *DataSet) GetDataSetAsJsonFiltered(omitBinaryVals, addReadableNames bool, tags []dicomtag.Tag) (string, error) {
	jObj, err := ds.GetDataSetAsJsonObjFiltered(omitBinaryVals, addReadableNames, tags)
	if err != nil {
		return "undefined", err
	}
	bytes, err := json.Marshal(jObj[0])
	return string(bytes), err
}
