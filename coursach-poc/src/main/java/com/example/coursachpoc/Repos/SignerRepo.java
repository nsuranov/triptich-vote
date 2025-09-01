package com.example.coursachpoc.Repos;

import com.example.coursachpoc.Entities.Signer;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.Optional;
import java.util.UUID;

public interface SignerRepo extends JpaRepository<Signer, UUID> {
    Optional<Signer> findSignerByPublicKey(String publicKey);
}
