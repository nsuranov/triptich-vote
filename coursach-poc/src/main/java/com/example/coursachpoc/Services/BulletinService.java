package com.example.coursachpoc.Services;

import com.example.coursachpoc.DTOs.BulletinCreateDTO;
import com.example.coursachpoc.Entities.Bulletin;
import com.example.coursachpoc.Entities.Candidate;
import com.example.coursachpoc.Repos.BulletinRepo;
import com.example.coursachpoc.Repos.CandidateRepo;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;
import org.springframework.web.server.ResponseStatusException;

import java.util.Map;

@Service
@RequiredArgsConstructor
public class BulletinService {
    private final BulletinRepo repo;
    private final CandidateRepo candidateRepo;
    private final RestTemplate rest;

    @Value("${verify.url:http://localhost:8088/verify}")
    private String verifyUrl;

    public void submit(BulletinCreateDTO dto) {
        // 1) верификация подписи
        Map<String,Object> req = Map.of(
                "message", dto.getCandidateId().toString(),
                "signatureB64", dto.getSignatureB64(),
                "ring", dto.getRing(),
                "n", dto.getN(),
                "m", dto.getM()
        );

        VerifyResponse res = rest.postForObject(verifyUrl, req, VerifyResponse.class);
        if (res == null || !Boolean.TRUE.equals(res.getOk())) {
            throw new ResponseStatusException(HttpStatus.BAD_REQUEST, "Invalid signature");
        }
        String uNum = res.getUNumber();

        // 2) проверка на повтор
        if (repo.existsByuNumber(uNum)) {
            throw new ResponseStatusException(HttpStatus.CONFLICT, "Duplicate vote");
        }

        // 3) сохранение
        Candidate cand = candidateRepo.findById(dto.getCandidateId())
                .orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND, "Candidate not found"));

        Bulletin b = new Bulletin();
        b.setUNumber(uNum);
        b.setRawData(dto.getSignatureB64());
        b.setCandidate(cand);

        repo.save(b);
    }

    @Data
    public static class VerifyResponse {
        private Boolean ok;
        @JsonProperty("uNumber")
        private String uNumber;
        private String error;
    }
}