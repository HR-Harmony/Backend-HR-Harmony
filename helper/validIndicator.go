package helper

import "hrsale/models"

// Function to validate KPI scores
func IsValidScore(kpiIndicator models.KPIIndicator) bool {
	// Check if scores are between 0 and 5
	if kpiIndicator.BddSellingSkill < 0 || kpiIndicator.BddSellingSkill > 5 {
		return false
	}
	if kpiIndicator.BddHandlingObjection < 0 || kpiIndicator.BddHandlingObjection > 5 {
		return false
	}
	if kpiIndicator.BddNegotiationSkill < 0 || kpiIndicator.BddNegotiationSkill > 5 {
		return false
	}
	if kpiIndicator.BddProposalDevelopment < 0 || kpiIndicator.BddProposalDevelopment > 5 {
		return false
	}
	if kpiIndicator.BddAfterSalesManagement < 0 || kpiIndicator.BddAfterSalesManagement > 5 {
		return false
	}
	if kpiIndicator.BddCustomerRelationshipManagement < 0 || kpiIndicator.BddCustomerRelationshipManagement > 5 {
		return false
	}
	if kpiIndicator.BddHubunganInterpersonal < 0 || kpiIndicator.BddHubunganInterpersonal > 5 {
		return false
	}
	if kpiIndicator.BddCommunicationSkill < 0 || kpiIndicator.BddCommunicationSkill > 5 {
		return false
	}
	if kpiIndicator.BsdProductKnowledge < 0 || kpiIndicator.BsdProductKnowledge > 5 {
		return false
	}
	if kpiIndicator.BsdProjectManagement < 0 || kpiIndicator.BsdProjectManagement > 5 {
		return false
	}
	if kpiIndicator.BsdDeliveringProceduresOrProcess < 0 || kpiIndicator.BsdDeliveringProceduresOrProcess > 5 {
		return false
	}
	if kpiIndicator.BsdCollaboratingProcess < 0 || kpiIndicator.BsdCollaboratingProcess > 5 {
		return false
	}
	if kpiIndicator.BsdCustomerSatisfaction < 0 || kpiIndicator.BsdCustomerSatisfaction > 5 {
		return false
	}
	if kpiIndicator.BsdSelfConfidence < 0 || kpiIndicator.BsdSelfConfidence > 5 {
		return false
	}
	if kpiIndicator.BsdEmphaty < 0 || kpiIndicator.BsdEmphaty > 5 {
		return false
	}
	if kpiIndicator.TidComputerLiteracy < 0 || kpiIndicator.TidComputerLiteracy > 5 {
		return false

	}
	if kpiIndicator.TidSystemDatabaseManagement < 0 || kpiIndicator.TidSystemDatabaseManagement > 5 {
		return false
	}
	if kpiIndicator.TidNetworkManagement < 0 || kpiIndicator.TidNetworkManagement > 5 {
		return false
	}
	if kpiIndicator.TidProgramDevelopment < 0 || kpiIndicator.TidProgramDevelopment > 5 {
		return false

	}
	if kpiIndicator.TidCodingManagement < 0 || kpiIndicator.TidCodingManagement > 5 {
		return false
	}
	if kpiIndicator.TidSystemAnalyze < 0 || kpiIndicator.TidSystemAnalyze > 5 {
		return false
	}
	if kpiIndicator.TidUserExperienceManagement < 0 || kpiIndicator.TidUserExperienceManagement > 5 {
		return false

	}
	if kpiIndicator.Creativity < 0 || kpiIndicator.Creativity > 5 {
		return false
	}
	if kpiIndicator.UltimateSpeed < 0 || kpiIndicator.UltimateSpeed > 5 {
		return false
	}
	if kpiIndicator.Reliable < 0 || kpiIndicator.Reliable > 5 {
		return false

	}

	if kpiIndicator.OpenMinded < 0 || kpiIndicator.OpenMinded > 5 {
		return false

	}
	if kpiIndicator.SuperiorService < 0 || kpiIndicator.SuperiorService > 5 {
		return false
	}
	if kpiIndicator.Integrity < 0 || kpiIndicator.Integrity > 5 {
		return false
	}
	if kpiIndicator.AgileEntrepreneur < 0 || kpiIndicator.AgileEntrepreneur > 5 {
		return false

	}
	if kpiIndicator.DayaTahanStress < 0 || kpiIndicator.DayaTahanStress > 5 {
		return false
	}
	if kpiIndicator.StabilitasEmosi < 0 || kpiIndicator.StabilitasEmosi > 5 {
		return false
	}
	if kpiIndicator.MotivasiBerprestasi < 0 || kpiIndicator.MotivasiBerprestasi > 5 {
		return false
	}

	if kpiIndicator.AttentionToDetail < 0 || kpiIndicator.AttentionToDetail > 5 {
		return false
	}
	if kpiIndicator.TimeManagement < 0 || kpiIndicator.TimeManagement > 5 {
		return false

	}
	if kpiIndicator.DisciplineExecution < 0 || kpiIndicator.DisciplineExecution > 5 {
		return false
	}
	if kpiIndicator.QualityOrientation < 0 || kpiIndicator.QualityOrientation > 5 {
		return false
	}

	return true
}

// Function to calculate total scores
func CalculateTotalScores(kpiIndicator models.KPIIndicator) float64 {
	return float64(kpiIndicator.BddSellingSkill + kpiIndicator.BddHandlingObjection + kpiIndicator.BddNegotiationSkill + kpiIndicator.BddProposalDevelopment + kpiIndicator.BddAfterSalesManagement + kpiIndicator.BddCustomerRelationshipManagement + kpiIndicator.BddHubunganInterpersonal + kpiIndicator.BddCommunicationSkill + kpiIndicator.BsdProductKnowledge + kpiIndicator.BsdProjectManagement + kpiIndicator.BsdDeliveringProceduresOrProcess + kpiIndicator.BsdCollaboratingProcess + kpiIndicator.BsdCustomerSatisfaction + kpiIndicator.BsdSelfConfidence + kpiIndicator.BsdEmphaty + kpiIndicator.TidComputerLiteracy + kpiIndicator.TidSystemDatabaseManagement + kpiIndicator.TidNetworkManagement + kpiIndicator.TidProgramDevelopment + kpiIndicator.TidCodingManagement + kpiIndicator.TidSystemAnalyze + kpiIndicator.TidUserExperienceManagement + kpiIndicator.Creativity + kpiIndicator.UltimateSpeed + kpiIndicator.Reliable + kpiIndicator.OpenMinded + kpiIndicator.SuperiorService + kpiIndicator.Integrity + kpiIndicator.AgileEntrepreneur + kpiIndicator.DayaTahanStress + kpiIndicator.StabilitasEmosi + kpiIndicator.MotivasiBerprestasi + kpiIndicator.AttentionToDetail + kpiIndicator.TimeManagement + kpiIndicator.DisciplineExecution + kpiIndicator.QualityOrientation)
}
