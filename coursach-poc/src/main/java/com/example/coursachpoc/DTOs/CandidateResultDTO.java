package com.example.coursachpoc.DTOs;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.UUID;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class CandidateResultDTO {
    private UUID id;
    private String fullname;
    private long votes;
}