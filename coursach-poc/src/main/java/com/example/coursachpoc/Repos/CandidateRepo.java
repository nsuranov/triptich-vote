package com.example.coursachpoc.Repos;

import com.example.coursachpoc.Entities.Candidate;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.UUID;

public interface CandidateRepo extends JpaRepository<Candidate, UUID> {
}
