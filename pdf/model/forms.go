/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

//
// High level manipulation of forms (AcroForm).
//

package model

import (
	"fmt"

	"github.com/unidoc/unidoc/common"
	. "github.com/unidoc/unidoc/pdf/core"
)

type PdfAcroForm struct {
	Fields          *[]*PdfField
	NeedAppearances PdfObject
	SigFlags        PdfObject
	CO              PdfObject
	DR              PdfObject
	DA              PdfObject
	Q               PdfObject
	XFA             PdfObject

	primitive *PdfIndirectObject
}

func NewPdfAcroForm() *PdfAcroForm {
	acroForm := &PdfAcroForm{}

	container := &PdfIndirectObject{}
	container.PdfObject = &PdfObjectDictionary{}

	acroForm.primitive = container
	return acroForm
}

// Used when loading forms from PDF files.
func (r *PdfReader) newPdfAcroFormFromDict(d *PdfObjectDictionary) (*PdfAcroForm, error) {
	acroForm := NewPdfAcroForm()

	if obj, has := (*d)["Fields"]; has {
		obj, err := r.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		fieldArray, ok := TraceToDirectObject(obj).(*PdfObjectArray)
		if !ok {
			return nil, fmt.Errorf("Fields not an array (%T)", obj)
		}

		fields := []*PdfField{}
		for _, obj := range *fieldArray {
			obj, err := r.traceToObject(obj)
			if err != nil {
				return nil, err
			}
			container, isIndirect := obj.(*PdfIndirectObject)
			if !isIndirect {
				return nil, fmt.Errorf("Field not in an indirect object")
			}
			field, err := r.newPdfFieldFromIndirectObject(container, nil)
			if err != nil {
				return nil, err
			}
			common.Log.Debug("AcroForm Field: %+v", *field)
			fields = append(fields, field)
		}
		acroForm.Fields = &fields
	}

	if obj, has := (*d)["NeedAppearances"]; has {
		acroForm.NeedAppearances = obj
	}
	if obj, has := (*d)["SigFlags"]; has {
		acroForm.SigFlags = obj
	}
	if obj, has := (*d)["CO"]; has {
		acroForm.CO = obj
	}
	if obj, has := (*d)["DR"]; has {
		acroForm.DR = obj
	}
	if obj, has := (*d)["DA"]; has {
		acroForm.DA = obj
	}
	if obj, has := (*d)["Q"]; has {
		acroForm.Q = obj
	}
	if obj, has := (*d)["XFA"]; has {
		acroForm.XFA = obj
	}

	return acroForm, nil
}

func (this *PdfAcroForm) GetContainingPdfObject() PdfObject {
	return this.primitive
}

func (this *PdfAcroForm) ToPdfObject() PdfObject {
	container := this.primitive
	dict := container.PdfObject.(*PdfObjectDictionary)

	if this.Fields != nil {
		arr := PdfObjectArray{}
		for _, field := range *this.Fields {
			arr = append(arr, field.ToPdfObject())
		}
		(*dict)["Fields"] = &arr
	}

	if this.NeedAppearances != nil {
		(*dict)["NeedAppearances"] = this.NeedAppearances
	}
	if this.SigFlags != nil {
		(*dict)["SigFlags"] = this.SigFlags
	}
	if this.CO != nil {
		(*dict)["CO"] = this.CO
	}
	if this.DR != nil {
		(*dict)["DR"] = this.DR
	}
	if this.DA != nil {
		(*dict)["DA"] = this.DA
	}
	if this.Q != nil {
		(*dict)["Q"] = this.Q
	}
	if this.XFA != nil {
		(*dict)["XFA"] = this.XFA
	}

	return container
}

// PdfField represents a field of an interactive form.
// Implements PdfModel interface.
type PdfField struct {
	FT     *PdfObjectName // field type
	Parent *PdfField
	// In a non-terminal field, the Kids array shall refer to field dictionaries that are immediate descendants of this field.
	// In a terminal field, the Kids array ordinarily shall refer to one or more separate widget annotations that are associated
	// with this field. However, if there is only one associated widget annotation, and its contents have been merged into the field
	// dictionary, Kids shall be omitted.
	KidsF []PdfModel // Kids can be array of other fields or widgets (PdfModel).
	KidsA []*PdfAnnotation
	T     PdfObject
	TU    PdfObject
	TM    PdfObject
	Ff    PdfObject // field flag
	V     PdfObject //value
	DV    PdfObject
	AA    PdfObject

	// Variable Text:
	DA PdfObject
	Q  PdfObject
	DS PdfObject
	RV PdfObject

	primitive *PdfIndirectObject
}

func NewPdfField() *PdfField {
	field := &PdfField{}

	container := &PdfIndirectObject{}
	container.PdfObject = &PdfObjectDictionary{}

	field.primitive = container
	return field
}

// Used when loading fields from PDF files.
func (r *PdfReader) newPdfFieldFromIndirectObject(container *PdfIndirectObject, parent *PdfField) (*PdfField, error) {
	d, isDict := container.PdfObject.(*PdfObjectDictionary)
	if !isDict {
		return nil, fmt.Errorf("Pdf Field indirect object not containing a dictionary")
	}

	field := NewPdfField()

	// Field type (required in terminal fields).
	// Can be /Btn /Tx /Ch /Sig
	// Required for a terminal field (inheritable).
	var err error
	if obj, has := (*d)["FT"]; has {
		obj, err = r.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		name, ok := obj.(*PdfObjectName)
		if !ok {
			return nil, fmt.Errorf("Invalid type of FT field (%T)", obj)
		}

		field.FT = name
	}

	// Partial field name (Optional)
	if obj, has := (*d)["T"]; has {
		field.T = obj
	}
	// Alternate description (Optional)
	if obj, has := (*d)["TU"]; has {
		field.TU = obj
	}
	// Mapping name (Optional)
	if obj, has := (*d)["TM"]; has {
		field.TM = obj
	}
	// Field flag. (Optional; inheritable)
	if obj, has := (*d)["Ff"]; has {
		field.Ff = obj
	}
	// Value (Optional; inheritable) - Various types depending on the field type.
	if obj, has := (*d)["V"]; has {
		field.V = obj
	}
	// Default value for reset (Optional; inheritable)
	if obj, has := (*d)["DV"]; has {
		field.DV = obj
	}
	// Additional actions dictionary (Optional)
	if obj, has := (*d)["AA"]; has {
		field.AA = obj
	}

	// Variable text:
	if obj, has := (*d)["DA"]; has {
		field.DA = obj
	}
	if obj, has := (*d)["Q"]; has {
		field.Q = obj
	}
	if obj, has := (*d)["DS"]; has {
		field.DS = obj
	}
	if obj, has := (*d)["RV"]; has {
		field.RV = obj
	}

	// In a non-terminal field, the Kids array shall refer to field dictionaries that are immediate descendants of this field.
	// In a terminal field, the Kids array ordinarily shall refer to one or more separate widget annotations that are associated
	// with this field. However, if there is only one associated widget annotation, and its contents have been merged into the field
	// dictionary, Kids shall be omitted.

	// Set ourself?
	if parent != nil {
		field.Parent = parent
	}

	// Has a merged-in widget annotation?
	if obj, has := (*d)["Subtype"]; has {
		obj, err = r.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		common.Log.Debug("Merged in annotation (%T)", obj)
		name, ok := obj.(*PdfObjectName)
		if !ok {
			return nil, fmt.Errorf("Invalid type of Subtype (%T)", obj)
		}
		if *name == "Widget" {
			// Is a merged field / widget dict.

			// Check if the annotation has already been loaded?
			// Most likely referenced to by a page...  Could be in either direction.
			// r.newPdfAnnotationFromIndirectObject acts as a caching mechanism.
			annot, err := r.newPdfAnnotationFromIndirectObject(container)
			if err != nil {
				return nil, err
			}
			widget, ok := annot.GetContext().(*PdfAnnotationWidget)
			if !ok {
				return nil, fmt.Errorf("Invalid widget")
			}

			widget.Parent = field.GetContainingPdfObject()
			field.KidsA = append(field.KidsA, annot)
			return field, nil
		}
	}

	if obj, has := (*d)["Kids"]; has {
		obj, err := r.traceToObject(obj)
		if err != nil {
			return nil, err
		}
		fieldArray, ok := TraceToDirectObject(obj).(*PdfObjectArray)
		if !ok {
			return nil, fmt.Errorf("Kids not an array (%T)", obj)
		}

		field.KidsF = []PdfModel{}
		for _, obj := range *fieldArray {
			obj, err := r.traceToObject(obj)
			if err != nil {
				return nil, err
			}

			container, isIndirect := obj.(*PdfIndirectObject)
			if !isIndirect {
				return nil, fmt.Errorf("Not an indirect object (form field)")
			}

			childField, err := r.newPdfFieldFromIndirectObject(container, field)
			if err != nil {
				return nil, err
			}

			field.KidsF = append(field.KidsF, childField)
		}
	}

	return field, nil
}

func (this *PdfField) GetContainingPdfObject() PdfObject {
	return this.primitive
}

// If Kids refer only to a single pdf widget annotation widget, then can merge it in.
// Currently not merging it in.
func (this *PdfField) ToPdfObject() PdfObject {
	container := this.primitive
	dict := container.PdfObject.(*PdfObjectDictionary)

	if this.Parent != nil {
		(*dict)["Parent"] = this.Parent.GetContainingPdfObject()
	}

	if this.KidsF != nil {
		// Create an array of the kids (fields or widgets).
		common.Log.Debug("KidsF: %+v", this.KidsF)
		arr := PdfObjectArray{}
		for _, child := range this.KidsF {
			arr = append(arr, child.ToPdfObject())
		}
		(*dict)["Kids"] = &arr
	}
	if this.KidsA != nil {
		common.Log.Debug("KidsA: %+v", this.KidsA)
		_, hasKids := (*dict)["Kids"].(*PdfObjectArray)
		if !hasKids {
			(*dict)["Kids"] = &PdfObjectArray{}
		}
		arr := (*dict)["Kids"].(*PdfObjectArray)
		for _, child := range this.KidsA {
			*arr = append(*arr, child.GetContext().ToPdfObject())
		}
	}

	if this.FT != nil {
		(*dict)["FT"] = this.FT
	}

	if this.T != nil {
		(*dict)["T"] = this.T
	}
	if this.TU != nil {
		(*dict)["TU"] = this.TU
	}
	if this.TM != nil {
		(*dict)["TM"] = this.TM
	}
	if this.Ff != nil {
		(*dict)["Ff"] = this.Ff
	}
	if this.V != nil {
		(*dict)["V"] = this.V
	}
	if this.DV != nil {
		(*dict)["DV"] = this.DV
	}
	if this.AA != nil {
		(*dict)["AA"] = this.AA
	}

	// Variable text:
	dict.SetIfNotNil("DA", this.DA)
	dict.SetIfNotNil("Q", this.Q)
	dict.SetIfNotNil("DS", this.DS)
	dict.SetIfNotNil("RV", this.RV)

	return container
}
