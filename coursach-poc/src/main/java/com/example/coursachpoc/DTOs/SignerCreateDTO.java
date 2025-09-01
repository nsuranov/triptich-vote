package com.example.coursachpoc.DTOs;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class SignerCreateDTO {
    private String fullName;
    private String publicKey;
}
