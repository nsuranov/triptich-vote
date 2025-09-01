package com.example.coursachpoc.Services;

import com.example.coursachpoc.DTOs.CandidateDTO;
import com.example.coursachpoc.DTOs.CandidateResultDTO;
import com.example.coursachpoc.Entities.Candidate;
import com.example.coursachpoc.Repos.CandidateRepo;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.Comparator;
import java.util.List;

@Service
@RequiredArgsConstructor
public class CandidateService {
    private final CandidateRepo candidateRepo;

    public CandidateDTO addCandidate(String candidateFullname) {
        Candidate candidate = new Candidate();
        candidate.setFullname(candidateFullname);
        candidate = candidateRepo.save(candidate);
        return new CandidateDTO(candidate.getId(), candidate.getFullname());
    }

    public List<CandidateDTO> getAllCandidates() {
        return candidateRepo.findAll().stream().map(this::castCandidateToDTO).toList();
    }

    private CandidateDTO castCandidateToDTO(Candidate candidate) {
        CandidateDTO candidateDTO = new CandidateDTO();
        candidateDTO.setId(candidate.getId());
        candidateDTO.setFullname(candidate.getFullname());
        return candidateDTO;
    }
    public List<CandidateResultDTO> getResults() {
        return candidateRepo.findAll().stream()
                .map(c -> new CandidateResultDTO(c.getId(), c.getFullname(),
                        c.getBulletins() != null ? c.getBulletins().size() : 0))
                .sorted(Comparator.comparingLong(CandidateResultDTO::getVotes).reversed())
                .toList();
    }
}
