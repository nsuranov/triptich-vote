package com.example.coursachpoc.DTOs;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.UUID;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class CandidateDTO {
    private UUID id;
    private String fullname;
}
