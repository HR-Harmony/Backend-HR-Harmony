package models

import "time"

type KPIIndicator struct {
	ID                                uint      `gorm:"primaryKey" json:"id"`
	Title                             string    `json:"title"`
	DesignationID                     uint      `json:"designation_id"`
	DesignationName                   string    `json:"designation_name"`
	AdminId                           uint      `json:"admin_id"`
	AdminName                         string    `json:"admin_name"`
	BddSellingSkill                   uint      `json:"bdd_selling_skill"`
	BddHandlingObjection              uint      `json:"bdd_handling_objection"`
	BddNegotiationSkill               uint      `json:"bdd_negotiation_skill"`
	BddProposalDevelopment            uint      `json:"bdd_proposal_development"`
	BddAfterSalesManagement           uint      `json:"bdd_after_sales_management"`
	BddCustomerRelationshipManagement uint      `json:"bdd_customer_relationship_management"`
	BddHubunganInterpersonal          uint      `json:"bdd_hubungan_interpersonal"`
	BddCommunicationSkill             uint      `json:"bdd_communication_skill"`
	BsdProductKnowledge               uint      `json:"bsd_product_knowledge"`
	BsdProjectManagement              uint      `json:"bsd_project_management"`
	BsdDeliveringProceduresOrProcess  uint      `json:"bsd_delivering_procedures_or_process"`
	BsdCollaboratingProcess           uint      `json:"bsd_collaborating_process"`
	BsdCustomerSatisfaction           uint      `json:"bsd_customer_satisfaction"`
	BsdSelfConfidence                 uint      `json:"bsd_self_confidence"`
	BsdEmphaty                        uint      `json:"bsd_emphaty"`
	TidComputerLiteracy               uint      `json:"tid_computer_literacy"`
	TidSystemDatabaseManagement       uint      `json:"tid_system_database_management"`
	TidNetworkManagement              uint      `json:"tid_network_management"`
	TidProgramDevelopment             uint      `json:"tid_program_development"`
	TidCodingManagement               uint      `json:"tid_coding_management"`
	TidSystemAnalyze                  uint      `json:"tid_system_analyze"`
	TidUserExperienceManagement       uint      `json:"tid_user_experience_management"`
	Creativity                        uint      `json:"creativity"`
	UltimateSpeed                     uint      `json:"ultimate_speed"`
	Reliable                          uint      `json:"reliable"`
	OpenMinded                        uint      `json:"open_minded"`
	SuperiorService                   uint      `json:"superior_service"`
	Integrity                         uint      `json:"integrity"`
	AgileEntrepreneur                 uint      `json:"agile_entrepreneur"`
	DayaTahanStress                   uint      `json:"daya_tahan_stress"`
	StabilitasEmosi                   uint      `json:"stabilitas_emosi"`
	MotivasiBerprestasi               uint      `json:"motivasi_berprestasi"`
	AttentionToDetail                 uint      `json:"attention_to_detail"`
	TimeManagement                    uint      `json:"time_management"`
	DisciplineExecution               uint      `json:"discipline_execution"`
	QualityOrientation                uint      `json:"quality_orientation"`
	Result                            float64   `json:"result"`
	CreatedAt                         time.Time `json:"created_at"`
	UpdatedAt                         time.Time `json:"updated_at"`
}

type KPAIndicator struct {
	ID                                uint      `gorm:"primaryKey" json:"id"`
	Title                             string    `json:"title"`
	EmployeeID                        uint      `json:"employee_id"`
	EmployeeName                      string    `json:"employee_name"`
	AdminId                           uint      `json:"admin_id"`
	AdminName                         string    `json:"admin_name"`
	BddSellingSkill                   uint      `json:"bdd_selling_skill"`
	BddHandlingObjection              uint      `json:"bdd_handling_objection"`
	BddNegotiationSkill               uint      `json:"bdd_negotiation_skill"`
	BddProposalDevelopment            uint      `json:"bdd_proposal_development"`
	BddAfterSalesManagement           uint      `json:"bdd_after_sales_management"`
	BddCustomerRelationshipManagement uint      `json:"bdd_customer_relationship_management"`
	BddHubunganInterpersonal          uint      `json:"bdd_hubungan_interpersonal"`
	BddCommunicationSkill             uint      `json:"bdd_communication_skill"`
	BsdProductKnowledge               uint      `json:"bsd_product_knowledge"`
	BsdProjectManagement              uint      `json:"bsd_project_management"`
	BsdDeliveringProceduresOrProcess  uint      `json:"bsd_delivering_procedures_or_process"`
	BsdCollaboratingProcess           uint      `json:"bsd_collaborating_process"`
	BsdCustomerSatisfaction           uint      `json:"bsd_customer_satisfaction"`
	BsdSelfConfidence                 uint      `json:"bsd_self_confidence"`
	BsdEmphaty                        uint      `json:"bsd_emphaty"`
	TidComputerLiteracy               uint      `json:"tid_computer_literacy"`
	TidSystemDatabaseManagement       uint      `json:"tid_system_database_management"`
	TidNetworkManagement              uint      `json:"tid_network_management"`
	TidProgramDevelopment             uint      `json:"tid_program_development"`
	TidCodingManagement               uint      `json:"tid_coding_management"`
	TidSystemAnalyze                  uint      `json:"tid_system_analyze"`
	TidUserExperienceManagement       uint      `json:"tid_user_experience_management"`
	Creativity                        uint      `json:"creativity"`
	UltimateSpeed                     uint      `json:"ultimate_speed"`
	Reliable                          uint      `json:"reliable"`
	OpenMinded                        uint      `json:"open_minded"`
	SuperiorService                   uint      `json:"superior_service"`
	Integrity                         uint      `json:"integrity"`
	AgileEntrepreneur                 uint      `json:"agile_entrepreneur"`
	DayaTahanStress                   uint      `json:"daya_tahan_stress"`
	StabilitasEmosi                   uint      `json:"stabilitas_emosi"`
	MotivasiBerprestasi               uint      `json:"motivasi_berprestasi"`
	AttentionToDetail                 uint      `json:"attention_to_detail"`
	TimeManagement                    uint      `json:"time_management"`
	DisciplineExecution               uint      `json:"discipline_execution"`
	QualityOrientation                uint      `json:"quality_orientation"`
	Result                            float64   `json:"result"`
	CreatedAt                         time.Time `json:"created_at"`
	UpdatedAt                         time.Time `json:"updated_at"`
}
