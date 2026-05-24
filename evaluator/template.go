// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package evaluator

/* NumericTemplate */

type NumericTemplate struct {
	ModelVisitor
}

/* StringTemplate */

type StringTemplate struct {
	ModelVisitor
}

/* TemplateFactory */

type TemplateMaker struct{

}

func (tm *TemplateMaker) Create() *TemplateMaker {
	return tm
}

func (tm *TemplateMaker) MakeNumericTemplateCase(vf VisitorFunc) INumericTemplate {
	return new(NumericTemplate).Create(vf)
}

func (tm *TemplateMaker) MakeStringTemplateCase(vf VisitorFunc) IStringTemplate {
	return new(StringTemplate).Create(vf)
}