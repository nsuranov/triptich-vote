package com.example.coursachpoc.DTOs;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;
import java.util.UUID;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class BulletinCreateDTO {
    private UUID candidateId;
    private String signatureB64;
    private List<String> ring;
    private int n;
    private int m;
}