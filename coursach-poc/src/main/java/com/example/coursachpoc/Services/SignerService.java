package com.example.coursachpoc.Services;

import com.example.coursachpoc.DTOs.RingDTO;
import com.example.coursachpoc.DTOs.SignerCreateDTO;
import com.example.coursachpoc.Entities.Signer;
import com.example.coursachpoc.Repos.SignerRepo;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.web.server.ResponseStatusException;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

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

    public RingDTO getRing(){
        List<Signer> signers = signerRepo.findAll();
        RingDTO ringDTO = new RingDTO();
        List<String>publicKeys = new ArrayList<>();
        for(Signer signer : signers){
            publicKeys.add(signer.getPublicKey());
        }
        ringDTO.setPublicKeys(publicKeys);
        return ringDTO;
    }
}
