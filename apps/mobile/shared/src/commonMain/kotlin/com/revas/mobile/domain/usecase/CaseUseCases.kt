package com.revas.mobile.domain.usecase

import com.revas.mobile.domain.model.CaseDetail
import com.revas.mobile.domain.model.MobileError
import com.revas.mobile.domain.repository.CaseRepository

class LoadCaseDetailUseCase(private val repository: CaseRepository) {
    suspend operator fun invoke(caseId: String): Outcome<CaseDetail> {
        if (caseId.isBlank()) {
            return Outcome.Failure(MobileError.Validation("ID kasus tidak valid"))
        }
        return repository.loadCase(caseId)
    }
}
