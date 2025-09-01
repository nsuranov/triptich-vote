package com.example.coursachpoc.Entities;

import jakarta.persistence.*;
import lombok.Getter;
import lombok.Setter;

import java.util.UUID;

@Entity
@Table(name = "bulletin")
@Getter
@Setter
public class Bulletin {
    @Id
    @GeneratedValue(strategy = GenerationType.AUTO)
    private UUID id;
    @Column(columnDefinition = "TEXT")
    private String uNumber;

    @Column(columnDefinition = "TEXT")
    private String rawData;


    @ManyToOne(fetch = FetchType.LAZY, optional = false)
    @JoinColumn(name = "candidate_id", nullable = false)
    private Candidate candidate;
}
