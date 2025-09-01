package com.example.coursachpoc.Entities;

import jakarta.persistence.*;
import lombok.Getter;
import lombok.Setter;

import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

@Entity
@Table(name = "candidate")
@Getter
@Setter
public class Candidate {
    @Id
    @GeneratedValue(strategy = GenerationType.AUTO)
    private UUID id;

    @Column(columnDefinition = "TEXT")
    private String fullname;

    @OneToMany(mappedBy = "candidate", cascade = CascadeType.ALL)
    private List<Bulletin> bulletins = new ArrayList<>();
}
