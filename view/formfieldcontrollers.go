package view

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/ungerik/go-start/model"
)

///////////////////////////////////////////////////////////////////////////////
// FormFieldController

// FormFieldController is a MVC controller for form fields.
type FormFieldController interface {
	// Supports returns if this controller supports a model with the given metaData.
	Supports(metaData *model.MetaData, form *Form) bool

	// NewInput creates a new form field input view for the model metaData.
	// Returns ErrFormFieldTypeNotSupported if the model is not supported.
	NewInput(withLabel bool, metaData *model.MetaData, form *Form) (input View, err error)

	// SetValue sets the value of the model from HTTP POST form data.
	// Returns ErrFormFieldTypeNotSupported if the model is not supported.
	SetValue(ctx *Context, metaData *model.MetaData, form *Form) error
}

///////////////////////////////////////////////////////////////////////////////
// ErrFormFieldTypeNotSupported

type ErrFormFieldTypeNotSupported struct {
	*model.MetaData
}

func (self ErrFormFieldTypeNotSupported) Error() string {
	return fmt.Sprintf("Type %s of form field %s not supported", self.Value.Type(), self.Selector())
}

///////////////////////////////////////////////////////////////////////////////
// FormFieldControllers

type FormFieldControllers []FormFieldController

func (self FormFieldControllers) Supports(metaData *model.MetaData, form *Form) bool {
	for _, c := range self {
		if c.Supports(metaData, form) {
			return true
		}
	}
	return false
}

func (self FormFieldControllers) NewInput(withLabel bool, metaData *model.MetaData, form *Form) (input View, err error) {
	for _, c := range self {
		if c.Supports(metaData, form) {
			return c.NewInput(withLabel, metaData, form)
		}
	}
	return nil, ErrFormFieldTypeNotSupported{metaData}
}

func (self FormFieldControllers) SetValue(ctx *Context, metaData *model.MetaData, form *Form) error {
	for _, c := range self {
		if c.Supports(metaData, form) {
			return c.SetValue(ctx, metaData, form)
		}
	}
	return ErrFormFieldTypeNotSupported{metaData}
}

///////////////////////////////////////////////////////////////////////////////
// modelValueControllerBase

type modelValueControllerBase struct{}

func (self modelValueControllerBase) SetValue(ctx *Context, metaData *model.MetaData, form *Form) error {
	value := metaData.Value.Addr().Interface().(model.Value)
	value.SetString(ctx.Request.FormValue(metaData.Selector()))
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// ModelStringController

type ModelStringController struct {
	modelValueControllerBase
}

func (self ModelStringController) Supports(metaData *model.MetaData, form *Form) bool {
	_, ok := metaData.Value.Addr().Interface().(*model.String)
	return ok
}

func (self ModelStringController) NewInput(withLabel bool, metaData *model.MetaData, form *Form) (input View, err error) {
	str := metaData.Value.Addr().Interface().(*model.String)
	textField := &TextField{
		Class:       form.FieldInputClass(metaData),
		Name:        metaData.Selector(),
		Text:        str.Get(),
		Size:        form.GetInputSize(metaData),
		Disabled:    form.IsFieldDisabled(metaData),
		Placeholder: form.InputFieldPlaceholder(metaData),
	}
	if maxlen, ok, _ := str.Maxlen(metaData); ok {
		textField.MaxLength = maxlen
		if maxlen < textField.Size {
			textField.Size = maxlen
		}
	}
	if withLabel {
		return AddStandardLabel(form, textField, metaData), nil
	}
	return textField, nil
}

///////////////////////////////////////////////////////////////////////////////
// ModelTextController

type ModelTextController struct {
	modelValueControllerBase
}

func (self ModelTextController) Supports(metaData *model.MetaData, form *Form) bool {
	_, ok := metaData.Value.Addr().Interface().(*model.Text)
	return ok
}

func (self ModelTextController) NewInput(withLabel bool, metaData *model.MetaData, form *Form) (input View, err error) {
	text := metaData.Value.Addr().Interface().(*model.Text)
	var cols int // will be zero if not available, which is OK
	if str, ok := metaData.Attrib(StructTagKey, "cols"); ok {
		cols, err = strconv.Atoi(str)
		if err != nil {
			panic("Error in StandardFormFieldFactory.NewInput(): " + err.Error())
		}
	}
	var rows int // will be zero if not available, which is OK
	if str, ok := metaData.Attrib(StructTagKey, "rows"); ok {
		rows, err = strconv.Atoi(str)
		if err != nil {
			panic("Error in StandardFormFieldFactory.NewInput(): " + err.Error())
		}
	}
	input = &TextArea{
		Class:       form.FieldInputClass(metaData),
		Name:        metaData.Selector(),
		Text:        text.Get(),
		Cols:        cols,
		Rows:        rows,
		Disabled:    form.IsFieldDisabled(metaData),
		Placeholder: form.InputFieldPlaceholder(metaData),
	}
	if withLabel {
		return AddStandardLabel(form, input, metaData), nil
	}
	return input, nil
}

///////////////////////////////////////////////////////////////////////////////
// ModelUrlController

type ModelUrlController struct {
	modelValueControllerBase
}

func (self ModelUrlController) Supports(metaData *model.MetaData, form *Form) bool {
	_, ok := metaData.Value.Addr().Interface().(*model.Url)
	return ok
}

func (self ModelUrlController) NewInput(withLabel bool, metaData *model.MetaData, form *Form) (input View, err error) {
	url := metaData.Value.Addr().Interface().(*model.Url)
	input = &TextField{
		Class:       form.FieldInputClass(metaData),
		Name:        metaData.Selector(),
		Text:        url.Get(),
		Size:        form.GetInputSize(metaData),
		Disabled:    form.IsFieldDisabled(metaData),
		Placeholder: form.InputFieldPlaceholder(metaData),
	}
	if withLabel {
		return AddStandardLabel(form, input, metaData), nil
	}
	return input, nil
}

///////////////////////////////////////////////////////////////////////////////
// ModelEmailController

type ModelEmailController struct {
	modelValueControllerBase
}

func (self ModelEmailController) Supports(metaData *model.MetaData, form *Form) bool {
	_, ok := metaData.Value.Addr().Interface().(*model.Email)
	return ok
}

func (self ModelEmailController) NewInput(withLabel bool, metaData *model.MetaData, form *Form) (input View, err error) {
	email := metaData.Value.Addr().Interface().(*model.Email)
	input = &TextField{
		Class:       form.FieldInputClass(metaData),
		Name:        metaData.Selector(),
		Type:        EmailTextField,
		Text:        email.Get(),
		Size:        form.GetInputSize(metaData),
		Disabled:    form.IsFieldDisabled(metaData),
		Placeholder: form.InputFieldPlaceholder(metaData),
	}
	if withLabel {
		return AddStandardLabel(form, input, metaData), nil
	}
	return input, nil
}

///////////////////////////////////////////////////////////////////////////////
// ModelPasswordController

type ModelPasswordController struct {
	modelValueControllerBase
}

func (self ModelPasswordController) Supports(metaData *model.MetaData, form *Form) bool {
	_, ok := metaData.Value.Addr().Interface().(*model.Password)
	return ok
}

func (self ModelPasswordController) NewInput(withLabel bool, metaData *model.MetaData, form *Form) (input View, err error) {
	password := metaData.Value.Addr().Interface().(*model.Password)
	textField := &TextField{
		Class:       form.FieldInputClass(metaData),
		Name:        metaData.Selector(),
		Type:        PasswordTextField,
		Text:        password.Get(),
		Size:        form.GetInputSize(metaData),
		Disabled:    form.IsFieldDisabled(metaData),
		Placeholder: form.InputFieldPlaceholder(metaData),
	}
	if maxlen, ok, _ := password.Maxlen(metaData); ok {
		textField.MaxLength = maxlen
		if maxlen < textField.Size {
			textField.Size = maxlen
		}
	}
	if withLabel {
		return AddStandardLabel(form, textField, metaData), nil
	}
	return textField, nil
}

///////////////////////////////////////////////////////////////////////////////
// ModelPhoneController

type ModelPhoneController struct {
	modelValueControllerBase
}

func (self ModelPhoneController) Supports(metaData *model.MetaData, form *Form) bool {
	_, ok := metaData.Value.Addr().Interface().(*model.Phone)
	return ok
}

func (self ModelPhoneController) NewInput(withLabel bool, metaData *model.MetaData, form *Form) (input View, err error) {
	phone := metaData.Value.Addr().Interface().(*model.Phone)
	input = &TextField{
		Class:       form.FieldInputClass(metaData),
		Name:        metaData.Selector(),
		Text:        phone.Get(),
		Size:        form.GetInputSize(metaData),
		Disabled:    form.IsFieldDisabled(metaData),
		Placeholder: form.InputFieldPlaceholder(metaData),
	}
	if withLabel {
		return AddStandardLabel(form, input, metaData), nil
	}
	return input, nil
}

///////////////////////////////////////////////////////////////////////////////
// ModelBoolController

type ModelBoolController struct{}

func (self ModelBoolController) Supports(metaData *model.MetaData, form *Form) bool {
	_, ok := metaData.Value.Addr().Interface().(*model.Bool)
	return ok
}

func (self ModelBoolController) NewInput(withLabel bool, metaData *model.MetaData, form *Form) (input View, err error) {
	b := metaData.Value.Addr().Interface().(*model.Bool)
	checkbox := &Checkbox{
		Class:    form.FieldInputClass(metaData),
		Name:     metaData.Selector(),
		Disabled: form.IsFieldDisabled(metaData),
		Checked:  b.Get(),
	}
	if withLabel {
		checkbox.Label = form.FieldLabel(metaData)
	}
	return checkbox, nil
}

func (self ModelBoolController) SetValue(ctx *Context, metaData *model.MetaData, form *Form) error {
	b := metaData.Value.Addr().Interface().(*model.Bool)
	b.Set(ctx.Request.FormValue(metaData.Selector()) != "")
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// ModelChoiceController

type ModelChoiceController struct {
	modelValueControllerBase
}

func (self ModelChoiceController) Supports(metaData *model.MetaData, form *Form) bool {
	_, ok := metaData.Value.Addr().Interface().(*model.Choice)
	return ok
}

func (self ModelChoiceController) NewInput(withLabel bool, metaData *model.MetaData, form *Form) (input View, err error) {
	choice := metaData.Value.Addr().Interface().(*model.Choice)
	options := choice.Options(metaData)
	if len(options) == 0 || options[0] != "" {
		options = append([]string{""}, options...)
	}
	input = &Select{
		Class:    form.FieldInputClass(metaData),
		Name:     metaData.Selector(),
		Model:    &StringsSelectModel{options, choice.Get()},
		Disabled: form.IsFieldDisabled(metaData),
		Size:     1,
	}
	if withLabel {
		return AddStandardLabel(form, input, metaData), nil
	}
	return input, nil
}

///////////////////////////////////////////////////////////////////////////////
// ModelMultipleChoiceController

type ModelMultipleChoiceController struct{}

func (self ModelMultipleChoiceController) Supports(metaData *model.MetaData, form *Form) bool {
	_, ok := metaData.Value.Addr().Interface().(*model.MultipleChoice)
	return ok
}

func (self ModelMultipleChoiceController) NewInput(withLabel bool, metaData *model.MetaData, form *Form) (input View, err error) {
	m := metaData.Value.Addr().Interface().(*model.MultipleChoice)
	options := m.Options(metaData)
	checkboxes := make(Views, len(options))
	for i, option := range options {
		checkboxes[i] = &Checkbox{
			Label:    option,
			Class:    form.FieldInputClass(metaData),
			Name:     fmt.Sprintf("%s_%d", metaData.Selector(), i),
			Disabled: form.IsFieldDisabled(metaData),
			Checked:  m.IsSet(option),
		}
	}
	if withLabel {
		return AddStandardLabel(form, checkboxes, metaData), nil
	}
	return checkboxes, nil
}

func (self ModelMultipleChoiceController) SetValue(ctx *Context, metaData *model.MetaData, form *Form) error {
	m := metaData.Value.Addr().Interface().(*model.MultipleChoice)
	options := m.Options(metaData)
	*m = nil
	for i, option := range options {
		name := fmt.Sprintf("%s_%d", metaData.Selector(), i)
		if ctx.Request.FormValue(name) != "" {
			*m = append(*m, option)
		}
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// ModelDynamicChoiceController

type ModelDynamicChoiceController struct {
	modelValueControllerBase
}

func (self ModelDynamicChoiceController) Supports(metaData *model.MetaData, form *Form) bool {
	_, ok := metaData.Value.Addr().Interface().(*model.DynamicChoice)
	return ok
}

func (self ModelDynamicChoiceController) NewInput(withLabel bool, metaData *model.MetaData, form *Form) (input View, err error) {
	choice := metaData.Value.Addr().Interface().(*model.DynamicChoice)
	options := choice.Options()
	index := choice.Index()
	if len(options) == 0 || options[0] != "" {
		options = append([]string{""}, options...)
		index++
	}
	input = &Select{
		Class:    form.FieldInputClass(metaData),
		Name:     metaData.Selector(),
		Model:    &IndexedStringsSelectModel{options, index},
		Disabled: form.IsFieldDisabled(metaData),
		Size:     1,
	}
	if withLabel {
		return AddStandardLabel(form, input, metaData), nil
	}
	return input, nil
}

///////////////////////////////////////////////////////////////////////////////
// ModelDateController

type ModelDateController struct {
	modelValueControllerBase
}

func (self ModelDateController) Supports(metaData *model.MetaData, form *Form) bool {
	_, ok := metaData.Value.Addr().Interface().(*model.Date)
	return ok
}

func (self ModelDateController) NewInput(withLabel bool, metaData *model.MetaData, form *Form) (input View, err error) {
	date := metaData.Value.Addr().Interface().(*model.Date)
	input = Views{
		HTML("(Format: " + model.DateFormat + ")<br/>"),
		&TextField{
			Class:       form.FieldInputClass(metaData),
			Name:        metaData.Selector(),
			Text:        date.Get(),
			Size:        len(model.DateFormat),
			Disabled:    form.IsFieldDisabled(metaData),
			Placeholder: form.InputFieldPlaceholder(metaData),
		},
	}
	if withLabel {
		return AddStandardLabel(form, input, metaData), nil
	}
	return input, nil
}

///////////////////////////////////////////////////////////////////////////////
// ModelDateTimeController

type ModelDateTimeController struct {
	modelValueControllerBase
}

func (self ModelDateTimeController) Supports(metaData *model.MetaData, form *Form) bool {
	_, ok := metaData.Value.Addr().Interface().(*model.DateTime)
	return ok
}

func (self ModelDateTimeController) NewInput(withLabel bool, metaData *model.MetaData, form *Form) (input View, err error) {
	dateTime := metaData.Value.Addr().Interface().(*model.DateTime)
	input = Views{
		HTML("(Format: " + model.DateTimeFormat + ")<br/>"),
		&TextField{
			Class:       form.FieldInputClass(metaData),
			Name:        metaData.Selector(),
			Text:        dateTime.Get(),
			Size:        len(model.DateTimeFormat),
			Disabled:    form.IsFieldDisabled(metaData),
			Placeholder: form.InputFieldPlaceholder(metaData),
		},
	}
	if withLabel {
		return AddStandardLabel(form, input, metaData), nil
	}
	return input, nil
}

///////////////////////////////////////////////////////////////////////////////
// ModelFloatController

type ModelFloatController struct {
	modelValueControllerBase
}

func (self ModelFloatController) Supports(metaData *model.MetaData, form *Form) bool {
	_, ok := metaData.Value.Addr().Interface().(*model.Float)
	return ok
}

func (self ModelFloatController) NewInput(withLabel bool, metaData *model.MetaData, form *Form) (input View, err error) {
	f := metaData.Value.Addr().Interface().(*model.Float)
	input = &TextField{
		Class:       form.FieldInputClass(metaData),
		Name:        metaData.Selector(),
		Text:        f.String(),
		Disabled:    form.IsFieldDisabled(metaData),
		Placeholder: form.InputFieldPlaceholder(metaData),
	}
	if withLabel {
		return AddStandardLabel(form, input, metaData), nil
	}
	return input, nil
}

///////////////////////////////////////////////////////////////////////////////
// ModelIntController

type ModelIntController struct {
	modelValueControllerBase
}

func (self ModelIntController) Supports(metaData *model.MetaData, form *Form) bool {
	_, ok := metaData.Value.Addr().Interface().(*model.Int)
	return ok
}

func (self ModelIntController) NewInput(withLabel bool, metaData *model.MetaData, form *Form) (input View, err error) {
	i := metaData.Value.Addr().Interface().(*model.Int)
	input = &TextField{
		Class:       form.FieldInputClass(metaData),
		Name:        metaData.Selector(),
		Text:        i.String(),
		Disabled:    form.IsFieldDisabled(metaData),
		Placeholder: form.InputFieldPlaceholder(metaData),
	}
	if withLabel {
		return AddStandardLabel(form, input, metaData), nil
	}
	return input, nil
}

///////////////////////////////////////////////////////////////////////////////
// ModelFileController

type ModelFileController struct{}

func (self ModelFileController) Supports(metaData *model.MetaData, form *Form) bool {
	_, ok := metaData.Value.Addr().Interface().(*model.File)
	return ok
}

func (self ModelFileController) NewInput(withLabel bool, metaData *model.MetaData, form *Form) (input View, err error) {
	input = &FileInput{
		Class:    form.FieldInputClass(metaData),
		Name:     metaData.Selector(),
		Disabled: form.IsFieldDisabled(metaData),
	}
	if withLabel {
		return AddStandardLabel(form, input, metaData), nil
	}
	return input, nil
}

func (self ModelFileController) SetValue(ctx *Context, metaData *model.MetaData, form *Form) error {
	f := metaData.Value.Addr().Interface().(*model.File)
	file, header, err := ctx.Request.FormFile(metaData.Selector())
	if err != nil {
		return err
	}
	defer file.Close()
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	f.Name = header.Filename
	f.Data = bytes
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// ModelBlobController

type ModelBlobController struct{}

func (self ModelBlobController) Supports(metaData *model.MetaData, form *Form) bool {
	_, ok := metaData.Value.Addr().Interface().(*model.Blob)
	return ok
}

func (self ModelBlobController) NewInput(withLabel bool, metaData *model.MetaData, form *Form) (input View, err error) {
	input = &FileInput{
		Class:    form.FieldInputClass(metaData),
		Name:     metaData.Selector(),
		Disabled: form.IsFieldDisabled(metaData),
	}
	if withLabel {
		return AddStandardLabel(form, input, metaData), nil
	}
	return input, nil
}

func (self ModelBlobController) SetValue(ctx *Context, metaData *model.MetaData, form *Form) error {
	b := metaData.Value.Addr().Interface().(*model.Blob)
	file, _, err := ctx.Request.FormFile(metaData.Selector())
	if err != nil {
		return err
	}
	defer file.Close()
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	b.Set(bytes)
	return nil
}
