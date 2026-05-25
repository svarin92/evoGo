// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package interfaces

type ITemplateVisitor interface {
	IVisitor
}

type (
	INumericTemplate interface {
		IVisitor
	}
	
	IStringTemplate interface {
		IVisitor
	}
)

type ITemplateMaker interface {
	MakeNumericTemplateCase(vf VisitorFunc) INumericTemplate
	MakeStringTemplateCase(vf VisitorFunc) IStringTemplate
}

// Generic formatting function (public interface).
type TemplateFunc func(IIndividual) bool

// TemplateFactory is a function that provides an ITemplateMaker.
type TemplateFactory func() ITemplateMaker
