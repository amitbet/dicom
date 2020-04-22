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
		val = el.Value
		// ----- sequences :-----
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

func GetDefaultMetadataTagFilter() []dicomtag.Tag {
	taglist := []dicomtag.Tag{}
	taglist = append(taglist, dicomtag.GetTagFromString("00020002"))
	taglist = append(taglist, dicomtag.GetTagFromString("00020003"))
	taglist = append(taglist, dicomtag.GetTagFromString("00020010"))
	taglist = append(taglist, dicomtag.GetTagFromString("00020012"))
	taglist = append(taglist, dicomtag.GetTagFromString("00020013"))
	taglist = append(taglist, dicomtag.GetTagFromString("00020016"))
	taglist = append(taglist, dicomtag.GetTagFromString("00080005"))
	taglist = append(taglist, dicomtag.GetTagFromString("00080008"))
	taglist = append(taglist, dicomtag.GetTagFromString("00080012"))
	taglist = append(taglist, dicomtag.GetTagFromString("00080013"))
	taglist = append(taglist, dicomtag.GetTagFromString("00080016"))
	taglist = append(taglist, dicomtag.GetTagFromString("00080018"))
	taglist = append(taglist, dicomtag.GetTagFromString("00080020"))
	taglist = append(taglist, dicomtag.GetTagFromString("00080021"))
	taglist = append(taglist, dicomtag.GetTagFromString("00080022"))
	taglist = append(taglist, dicomtag.GetTagFromString("00080023"))
	taglist = append(taglist, dicomtag.GetTagFromString("00080030"))
	taglist = append(taglist, dicomtag.GetTagFromString("00080031"))
	taglist = append(taglist, dicomtag.GetTagFromString("00080032"))
	taglist = append(taglist, dicomtag.GetTagFromString("00080033"))
	taglist = append(taglist, dicomtag.GetTagFromString("00080050"))
	taglist = append(taglist, dicomtag.GetTagFromString("00080054"))
	taglist = append(taglist, dicomtag.GetTagFromString("00080060"))
	taglist = append(taglist, dicomtag.GetTagFromString("00080070"))
	taglist = append(taglist, dicomtag.GetTagFromString("00080080"))
	taglist = append(taglist, dicomtag.GetTagFromString("00080090"))
	taglist = append(taglist, dicomtag.GetTagFromString("00081010"))
	taglist = append(taglist, dicomtag.GetTagFromString("00081030"))
	taglist = append(taglist, dicomtag.GetTagFromString("0008103E"))
	taglist = append(taglist, dicomtag.GetTagFromString("00081060"))
	taglist = append(taglist, dicomtag.GetTagFromString("00081070"))
	taglist = append(taglist, dicomtag.GetTagFromString("00081090"))
	taglist = append(taglist, dicomtag.GetTagFromString("00090010"))
	taglist = append(taglist, dicomtag.GetTagFromString("00100010"))
	taglist = append(taglist, dicomtag.GetTagFromString("00100020"))
	taglist = append(taglist, dicomtag.GetTagFromString("00100030"))
	taglist = append(taglist, dicomtag.GetTagFromString("00100040"))
	taglist = append(taglist, dicomtag.GetTagFromString("00101001"))
	taglist = append(taglist, dicomtag.GetTagFromString("00101010"))
	taglist = append(taglist, dicomtag.GetTagFromString("001021B0"))
	taglist = append(taglist, dicomtag.GetTagFromString("00180010"))
	taglist = append(taglist, dicomtag.GetTagFromString("00180022"))
	taglist = append(taglist, dicomtag.GetTagFromString("00180050"))
	taglist = append(taglist, dicomtag.GetTagFromString("00180060"))
	taglist = append(taglist, dicomtag.GetTagFromString("00180090"))
	taglist = append(taglist, dicomtag.GetTagFromString("00181020"))
	taglist = append(taglist, dicomtag.GetTagFromString("00181030"))
	taglist = append(taglist, dicomtag.GetTagFromString("00181040"))
	taglist = append(taglist, dicomtag.GetTagFromString("00181100"))
	taglist = append(taglist, dicomtag.GetTagFromString("00181110"))
	taglist = append(taglist, dicomtag.GetTagFromString("00181111"))
	taglist = append(taglist, dicomtag.GetTagFromString("00181120"))
	taglist = append(taglist, dicomtag.GetTagFromString("00181130"))
	taglist = append(taglist, dicomtag.GetTagFromString("00181140"))
	taglist = append(taglist, dicomtag.GetTagFromString("00181150"))
	taglist = append(taglist, dicomtag.GetTagFromString("00181151"))
	taglist = append(taglist, dicomtag.GetTagFromString("00181152"))
	taglist = append(taglist, dicomtag.GetTagFromString("00181170"))
	taglist = append(taglist, dicomtag.GetTagFromString("00181190"))
	taglist = append(taglist, dicomtag.GetTagFromString("00181210"))
	taglist = append(taglist, dicomtag.GetTagFromString("00185100"))
	taglist = append(taglist, dicomtag.GetTagFromString("00190010"))
	taglist = append(taglist, dicomtag.GetTagFromString("0020000D"))
	taglist = append(taglist, dicomtag.GetTagFromString("0020000E"))
	taglist = append(taglist, dicomtag.GetTagFromString("00200010"))
	taglist = append(taglist, dicomtag.GetTagFromString("00200011"))
	taglist = append(taglist, dicomtag.GetTagFromString("00200012"))
	taglist = append(taglist, dicomtag.GetTagFromString("00200013"))
	taglist = append(taglist, dicomtag.GetTagFromString("00200032"))
	taglist = append(taglist, dicomtag.GetTagFromString("00200037"))
	taglist = append(taglist, dicomtag.GetTagFromString("00200052"))
	taglist = append(taglist, dicomtag.GetTagFromString("00201040"))
	taglist = append(taglist, dicomtag.GetTagFromString("00201041"))
	taglist = append(taglist, dicomtag.GetTagFromString("00210010"))
	taglist = append(taglist, dicomtag.GetTagFromString("00230010"))
	taglist = append(taglist, dicomtag.GetTagFromString("00270010"))
	taglist = append(taglist, dicomtag.GetTagFromString("00280002"))
	taglist = append(taglist, dicomtag.GetTagFromString("00280004"))
	taglist = append(taglist, dicomtag.GetTagFromString("00280010"))
	taglist = append(taglist, dicomtag.GetTagFromString("00280011"))
	taglist = append(taglist, dicomtag.GetTagFromString("00280030"))
	taglist = append(taglist, dicomtag.GetTagFromString("00280100"))
	taglist = append(taglist, dicomtag.GetTagFromString("00280101"))
	taglist = append(taglist, dicomtag.GetTagFromString("00280102"))
	taglist = append(taglist, dicomtag.GetTagFromString("00280103"))
	taglist = append(taglist, dicomtag.GetTagFromString("00280120"))
	taglist = append(taglist, dicomtag.GetTagFromString("00281050"))
	taglist = append(taglist, dicomtag.GetTagFromString("00281051"))
	taglist = append(taglist, dicomtag.GetTagFromString("00281052"))
	taglist = append(taglist, dicomtag.GetTagFromString("00281053"))
	taglist = append(taglist, dicomtag.GetTagFromString("00430010"))
	taglist = append(taglist, dicomtag.GetTagFromString("00450010"))
	taglist = append(taglist, dicomtag.GetTagFromString("00490010"))
	return taglist
}
