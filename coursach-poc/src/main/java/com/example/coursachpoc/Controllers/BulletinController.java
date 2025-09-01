package com.example.coursachpoc.Controllers;

import com.example.coursachpoc.DTOs.BulletinCreateDTO;
import com.example.coursachpoc.Services.BulletinService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController()
@RequestMapping("/api/bulletin")
@RequiredArgsConstructor
public class BulletinController {
    private final BulletinService bulletinService;

    @PostMapping
    public ResponseEntity<Void> submit(@RequestBody BulletinCreateDTO dto) {
        bulletinService.submit(dto);
        return ResponseEntity.ok().build();
    }
}
