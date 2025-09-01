package com.example.coursachpoc.Repos;

import com.example.coursachpoc.Entities.Bulletin;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.UUID;

public interface BulletinRepo extends JpaRepository<Bulletin, UUID> {
    boolean existsByuNumber(String uNumber);
}
