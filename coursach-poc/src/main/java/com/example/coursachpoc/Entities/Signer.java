package com.example.coursachpoc.Entities;

import jakarta.persistence.*;
import lombok.Getter;
import lombok.Setter;

import java.util.UUID;

@Entity
@Table(name = "signer")
@Getter
@Setter
public class Signer {
    @Id
    @GeneratedValue(strategy = GenerationType.AUTO)
    private UUID id;
    @Column(columnDefinition = "TEXT")
    private String fullname;
    @Column(columnDefinition = "TEXT")
    private String publicKey;
}
