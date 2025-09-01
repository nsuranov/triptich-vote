package com.example.coursachpoc.Controllers;

import com.example.coursachpoc.DTOs.CandidateDTO;
import com.example.coursachpoc.DTOs.CandidateResultDTO;
import com.example.coursachpoc.Entities.Candidate;
import com.example.coursachpoc.Services.CandidateService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController()
@RequestMapping("/api/candidate")
@RequiredArgsConstructor
public class CandidateController {
    private final CandidateService candidateService;

    @PostMapping()
    public ResponseEntity<CandidateDTO> addCandidate(@RequestParam String candidateName) {
        return ResponseEntity.ok(candidateService.addCandidate(candidateName));
    }

    @GetMapping()
    public ResponseEntity<List<CandidateDTO>> getAllCandidates() {
        return ResponseEntity.ok(candidateService.getAllCandidates());
    }

    @GetMapping("/results")
    public ResponseEntity<List<CandidateResultDTO>> getResults() {
        return ResponseEntity.ok(candidateService.getResults());
    }
}
