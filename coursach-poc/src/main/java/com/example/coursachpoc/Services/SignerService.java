package com.example.coursachpoc.Services;

import com.example.coursachpoc.DTOs.RingDTO;
import com.example.coursachpoc.DTOs.SignerCreateDTO;
import com.example.coursachpoc.Entities.Signer;
import com.example.coursachpoc.Repos.SignerRepo;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.web.server.ResponseStatusException;

import java.util.*;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
public class SignerService {
    private final SignerRepo signerRepo;

    public void createSigner(SignerCreateDTO signerDTO){
        Optional<Signer>  signerOptional = signerRepo.findSignerByPublicKey(signerDTO.getPublicKey());
        if(signerOptional.isPresent()){
            throw new ResponseStatusException(HttpStatus.BAD_REQUEST,"Signer Already Exists");
        }
        Signer signer = new Signer();
        signer.setFullname(signerDTO.getFullName());
        signer.setPublicKey(signerDTO.getPublicKey());
        signerRepo.save(signer);
    }

    public RingDTO getRing(Integer exp) {
        List<Signer> signers = signerRepo.findAll();
        int available = signers.size();

        int requested;
        if (exp == -1) {
            // ищем максимальную степень двойки, которая помещается в available
            requested = available;
        } else {
            requested = (int) Math.pow(2, exp);
        }

        int count = 1;
        int realExp = 0;
        while (count * 2 <= requested && count * 2 <= available) {
            count *= 2;
            realExp++;
        }

        Collections.shuffle(signers);
        List<String> publicKeys = signers.stream()
                .limit(count)
                .map(Signer::getPublicKey)
                .collect(Collectors.toList());

        return new RingDTO(publicKeys, (long) count, (long) realExp, 2L);
    }


}
