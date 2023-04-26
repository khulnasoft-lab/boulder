// Code generated by "stringer -type=FeatureFlag"; DO NOT EDIT.

package features

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[unused-0]
	_ = x[StoreRevokerInfo-1]
	_ = x[ROCSPStage6-2]
	_ = x[ROCSPStage7-3]
	_ = x[CAAValidationMethods-4]
	_ = x[CAAAccountURI-5]
	_ = x[EnforceMultiVA-6]
	_ = x[MultiVAFullResults-7]
	_ = x[ECDSAForAll-8]
	_ = x[ServeRenewalInfo-9]
	_ = x[AllowUnrecognizedFeatures-10]
	_ = x[ExpirationMailerUsesJoin-11]
	_ = x[CertCheckerChecksValidations-12]
	_ = x[CertCheckerRequiresValidations-13]
	_ = x[AsyncFinalize-14]
	_ = x[RequireCommonName-15]
	_ = x[StoreLintingCertificateInsteadOfPrecertificate-16]
}

const _FeatureFlag_name = "unusedStoreRevokerInfoROCSPStage6ROCSPStage7CAAValidationMethodsCAAAccountURIEnforceMultiVAMultiVAFullResultsECDSAForAllServeRenewalInfoAllowUnrecognizedFeaturesExpirationMailerUsesJoinCertCheckerChecksValidationsCertCheckerRequiresValidationsAsyncFinalizeRequireCommonNameStoreLintingCertificateInsteadOfPrecertificate"

var _FeatureFlag_index = [...]uint16{0, 6, 22, 33, 44, 64, 77, 91, 109, 120, 136, 161, 185, 213, 243, 256, 273, 319}

func (i FeatureFlag) String() string {
	if i < 0 || i >= FeatureFlag(len(_FeatureFlag_index)-1) {
		return "FeatureFlag(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _FeatureFlag_name[_FeatureFlag_index[i]:_FeatureFlag_index[i+1]]
}
