package com.example.coursachpoc.Controllers;

import com.example.coursachpoc.DTOs.RingDTO;
import com.example.coursachpoc.DTOs.SignerCreateDTO;
import com.example.coursachpoc.Services.SignerService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController()
@RequestMapping("/api/signer")
@RequiredArgsConstructor
public class SignerController {
    private final SignerService signerService;
    @PostMapping
    public ResponseEntity<Void> register(@RequestBody SignerCreateDTO signerCreateDTO) {
        signerService.createSigner(signerCreateDTO);
        return ResponseEntity.ok().build();
    }

    @GetMapping("/ring")
    public ResponseEntity<RingDTO> getRing() {
        return ResponseEntity.ok().body(signerService.getRing(-1));
    }
}